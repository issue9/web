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
	"github.com/issue9/sliceutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/content"
	"github.com/issue9/web/serialization"
	"github.com/issue9/web/service"
)

type contextKey int

const (
	contextKeyServer contextKey = iota
	contextKeyContext
)

// Server 提供了用于构建 Context 对象的基本数据
type Server struct {
	name       string
	version    string
	logs       *logs.Logs
	fs         fs.FS
	httpServer *http.Server
	vars       *sync.Map

	closed      chan struct{} // 当 Close 延时关闭时，通过此事件确定 Close() 的退出时机。
	closeEvents []func() error

	// middleware
	groups        *group.Groups
	compress      *compress.Compress
	errorHandlers *errorhandler.ErrorHandler
	routers       map[string]*Router

	// locale
	location *time.Location

	cache    cache.Cache
	uptime   time.Time
	modules  []*Module
	content  *content.Content
	services *service.Manager
	events   map[string]events.Eventer
}

// New 返回 *Server 实例
//
// name, version 表示服务的名称和版本号；
// o 指定了初始化 Server 一些非必要参数。在传递给 New 之后，再对其值进行改变，是无效的。
func New(name, version string, o *Options) (*Server, error) {
	o, err := o.sanitize()
	if err != nil {
		return nil, err
	}

	srv := &Server{
		name:       name,
		version:    version,
		logs:       o.Logs,
		fs:         o.FS,
		httpServer: o.httpServer,
		vars:       &sync.Map{},

		closed:      make(chan struct{}, 1),
		closeEvents: make([]func() error, 0, 3),

		groups:        o.groups,
		compress:      compress.Classic(o.Logs.ERROR(), o.IgnoreCompressTypes...),
		errorHandlers: errorhandler.New(),
		routers:       make(map[string]*Router, 3),

		location: o.Location,

		cache:    o.Cache,
		modules:  make([]*Module, 0, 20),
		uptime:   time.Now(),
		content:  content.New(o.ResultBuilder, o.Locale, o.Tag),
		services: service.NewManager(o.Logs, o.Location),
		events:   make(map[string]events.Eventer, 5),
	}
	srv.httpServer.Handler = srv.groups

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
	srv.MuxGroups().Middlewares().
		Append(recoverFunc.Middleware).      // 在最外层，防止协程 panic，崩了整个进程。
		Append(srv.compress.Middleware).     // srv.compress 会输出专有报头，所以应该在所有的输出内容之前。
		Append(srv.errorHandlers.Middleware) // errorHandler 依赖 recovery，必须要在 recovery 之后。

	// 加载插件，需要放在 srv 初始化完成之后。
	if o.Plugins != "" {
		if err := srv.loadPlugins(o.Plugins); err != nil {
			return nil, err
		}
	}

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

// Get 返回指定键名的值
func (srv *Server) Get(key interface{}) (interface{}, bool) { return srv.vars.Load(key) }

// Set 保存指定键名的值
func (srv *Server) Set(key, val interface{}) { srv.vars.Store(key, val) }

// Delete 删除指定键名的值
func (srv *Server) Delete(key interface{}) { srv.vars.Delete(key) }

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

// Services 返回服务内容的管理接口
func (srv *Server) Services() *service.Manager { return srv.services }

// Serve 启动服务
func (srv *Server) Serve(tag string, serve bool) (err error) {
	if err := srv.initModules(tag); err != nil {
		return err
	}
	if !serve {
		return nil
	}

	srv.Services().Run()

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
		srv.Logs().Flush()
		srv.closed <- struct{}{}
	}()

	srv.Services().Stop() // 先关闭服务
	srv.callCloseEvents()

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

func (srv *Server) callCloseEvents() {
	sliceutil.Reverse(srv.closeEvents)
	for _, event := range srv.closeEvents {
		if err := event(); err != nil {
			srv.Logs().Error(err)
		}
	}
}

// RegisterOnClose 注册关闭服务时的行为
//
// 按注册顺序反向调用
func (srv *Server) RegisterOnClose(f ...func() error) {
	srv.closeEvents = append(srv.closeEvents, f...)
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
func (t *Tag) Server() *Server { return t.Module().Server() }

// Mimetypes 返回用于序列化 web 内容的操作接口
func (srv *Server) Mimetypes() *serialization.Mimetypes { return srv.content.Mimetypes() }

// Files 返回用于序列化文件内容的操作接口
func (srv *Server) Files() *serialization.Files { return srv.content.Files() }

// Locale 返回用于序列化文件内容的操作接口
func (srv *Server) Locale() *serialization.Locale { return srv.content.Locale() }

func (srv *Server) LocalePrinter() *message.Printer { return srv.content.LocalePrinter() }

// Tag 返回默认的语言标签
func (srv *Server) Tag() language.Tag { return srv.content.Tag() }

// Results 返回在指定语言下的所有代码以及关联的描述信息
func (srv *Server) Results(p *message.Printer) map[int]string { return srv.content.Results(p) }

// AddResult 添加错误代码与关联的描述信息
func (srv *Server) AddResult(status, code int, key message.Reference, v ...interface{}) {
	srv.content.AddResult(status, code, key, v...)
}

// Result 返回指定代码在指定语言的错误描述信息
func (srv *Server) Result(p *message.Printer, code int, fields content.ResultFields) content.Result {
	return srv.content.Result(p, code, fields)
}
