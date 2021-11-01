// SPDX-License-Identifier: MIT

// Package server web 服务管理
package server

import (
	"context"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/events"
	"github.com/issue9/logs/v3"
	"github.com/issue9/middleware/v5/compress"
	"github.com/issue9/middleware/v5/errorhandler"
	"github.com/issue9/mux/v5/group"
	"github.com/issue9/scheduled"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/serialization"
)

type contextKey int

const (
	contextKeyServer contextKey = iota
	contextKeyContext
)

// DefaultMimetype 默认的媒体类型
//
// 在不能获取输入和输出的媒体类型时，会采用此值作为其默认值。
const DefaultMimetype = "application/octet-stream"

// DefaultCharset 默认的字符集
const DefaultCharset = "utf-8"

// Server 提供 HTTP 服务
type Server struct {
	name       string
	version    string
	logs       *logs.Logs
	fs         fs.FS
	httpServer *http.Server
	vars       *sync.Map
	mimetypes  *serialization.Mimetypes
	cache      cache.Cache
	uptime     time.Time
	modules    []*Module
	events     map[string]events.Eventer
	serving    bool

	closed chan struct{} // 当 Close 延时关闭时，通过此事件确定 Close() 的退出时机。

	// service
	services  []*Service
	scheduled *scheduled.Server

	// middleware
	group         *group.Group
	compress      *compress.Compress
	errorHandlers *errorhandler.ErrorHandler
	routers       map[string]*Router

	// result
	resultMessages map[string]*resultMessage
	resultBuilder  BuildResultFunc

	// locale
	location      *time.Location
	locale        *serialization.Locale
	tag           language.Tag
	localePrinter *message.Printer
}

// New 返回 *Server 实例
//
// name, version 表示服务的名称和版本号；
// o 指定了初始化 Server 一些非必要参数。在传递给 New 之后，再对其值进行改变，是无效的。
func New(name, version string, o *Options) (*Server, error) {
	if o == nil {
		o = &Options{}
	}
	if err := o.sanitize(); err != nil {
		return nil, err
	}

	srv := &Server{
		name:       name,
		version:    version,
		logs:       o.Logs,
		fs:         o.FS,
		httpServer: o.httpServer,
		vars:       &sync.Map{},
		mimetypes:  serialization.NewMimetypes(10),
		cache:      o.Cache,
		modules:    make([]*Module, 0, 20),
		uptime:     time.Now(),
		events:     make(map[string]events.Eventer, 5),

		closed: make(chan struct{}, 1),

		// service
		services:  make([]*Service, 0, 100),
		scheduled: scheduled.NewServer(o.Location),

		// middleware
		group:         o.group,
		compress:      compress.Classic(o.Logs.ERROR(), o.IgnoreCompressTypes...),
		errorHandlers: errorhandler.New(),
		routers:       make(map[string]*Router, 3),

		// result
		resultMessages: make(map[string]*resultMessage, 20),
		resultBuilder:  o.ResultBuilder,

		// locale
		location:      o.Location,
		locale:        o.locale,
		tag:           o.Tag,
		localePrinter: o.locale.Printer(o.Tag),
	}

	srv.httpServer.Handler = srv.group
	if srv.httpServer.BaseContext == nil {
		srv.httpServer.BaseContext = func(n net.Listener) context.Context {
			return context.WithValue(context.Background(), contextKeyServer, srv)
		}
	} else {
		ctx := srv.httpServer.BaseContext
		srv.httpServer.BaseContext = func(n net.Listener) context.Context {
			return context.WithValue(ctx(n), contextKeyServer, srv)
		}
	}

	recoverFunc := srv.errorHandlers.Recovery(o.Recovery)
	srv.MuxGroup().Middlewares().
		Append(recoverFunc.Middleware).      // 在最外层，防止协程 panic，崩了整个进程。
		Append(srv.compress.Middleware).     // srv.compress 会输出专有报头，所以应该在所有的输出内容之前。
		Append(srv.errorHandlers.Middleware) // errorHandler 依赖 recovery，必须要在 recovery 之后。

	return srv, nil
}

// GetServer 从请求中获取 *Server 实例
//
// r 必须得是由 Server 生成的，否则会 panic。
func GetServer(r *http.Request) *Server {
	if v := r.Context().Value(contextKeyServer); v != nil {
		return v.(*Server)
	}
	panic("无法从 http.Request.Context() 中获取 contentKeyServer 对应的值")
}

// Name 应用的名称
func (srv *Server) Name() string { return srv.name }

// Version 应用的版本
func (srv *Server) Version() string { return srv.version }

// Open 实现 fs.FS 接口
func (srv *Server) Open(name string) (fs.File, error) { return srv.fs.Open(name) }

// Vars 操纵共享变量的接口
func (srv *Server) Vars() *sync.Map { return srv.vars }

// Location 指定服务器的时区信息
func (srv *Server) Location() *time.Location { return srv.location }

// Logs 返回关联的 logs.Logs 实例
func (srv *Server) Logs() *logs.Logs { return srv.logs }

// Cache 返回缓存的相关接口
func (srv *Server) Cache() cache.Cache { return srv.cache }

// Uptime 当前服务的运行时间
func (srv *Server) Uptime() time.Time { return srv.uptime }

// Now 返回当前时间
//
// 与 time.Now() 的区别在于 Now() 基于当前时区
func (srv *Server) Now() time.Time { return time.Now().In(srv.Location()) }

// ParseTime 分析基于当前时区的时间
func (srv *Server) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, srv.Location())
}

// Jobs 返回所有的计划任务
func (srv *Server) Jobs() []*ScheduledJob { return srv.scheduled.Jobs() }

// Serve 启动服务
//
// serve 如果为空，表示不启动 HTTP 服务，仅执行向 action 注册的函数。
func (srv *Server) Serve(serve bool, action string) (err error) {
	if err := srv.initModules(false, action); err != nil {
		return err
	}
	if !serve {
		return nil
	}

	srv.runServices()

	// 在 Serve 中关闭服务，而不是 Close。 这样可以保证在所有的请求关闭之后执行。
	defer func() {
		srv.stopServices()
		err = errs.Merge(err, srv.initModules(true, action))
		err = errs.Merge(err, srv.Logs().Flush())
	}()

	srv.serving = true

	cfg := srv.httpServer.TLSConfig
	if cfg != nil && (cfg.GetCertificate != nil || len(cfg.Certificates) > 0) {
		err = srv.httpServer.ListenAndServeTLS("", "")
	} else {
		err = srv.httpServer.ListenAndServe()
	}

	// 由 Shutdown() 或 Close() 主动触发的关闭事件，才需要等待其执行完成，
	// 其它错误直接返回，否则一些内部错误会永远卡在此处无法返回。
	if errors.Is(err, http.ErrServerClosed) {
		<-srv.closed
	}
	return err
}

// Close 关闭服务
func (srv *Server) Close(shutdownTimeout time.Duration) error {
	defer func() {
		srv.closed <- struct{}{}
	}()

	srv.serving = false

	if shutdownTimeout == 0 {
		return srv.httpServer.Close()
	}

	c, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.httpServer.Shutdown(c); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return nil
}

// DisableCompression 是否禁用压缩功能
func (srv *Server) DisableCompression(disable bool) { srv.compress.Enable = !disable }

// SetErrorHandle 设置指定状态码页面的处理函数
//
// 如果状态码已经存在处理函数，则修改，否则就添加。
func (srv *Server) SetErrorHandle(h errorhandler.HandleFunc, status ...int) {
	srv.errorHandlers.Set(h, status...)
}

// Server 获取关联的 Server 实例
func (ctx *Context) Server() *Server { return ctx.server }

// Server 获取关联的 Server 实例
func (m *Module) Server() *Server { return m.srv }

// Server 获取关联的 Server 实例
func (t *Action) Server() *Server { return t.Module().Server() }

// Mimetypes 返回用于序列化 web 内容的操作接口
func (srv *Server) Mimetypes() *serialization.Mimetypes { return srv.mimetypes }

// Files 返回用于序列化文件内容的操作接口
func (srv *Server) Files() *serialization.Files { return srv.Locale().Files() }

// Locale 返回用于序列化文件内容的操作接口
func (srv *Server) Locale() *serialization.Locale { return srv.locale }

func (srv *Server) LocalePrinter() *message.Printer { return srv.localePrinter }

// Tag 返回默认的语言标签
func (srv *Server) Tag() language.Tag { return srv.tag }

// Serving 是否处于服务状态
func (srv *Server) Serving() bool { return srv.serving }
