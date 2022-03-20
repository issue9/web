// SPDX-License-Identifier: MIT

// Package server web 服务管理
package server

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v3"
	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/params"
	"github.com/issue9/scheduled"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/serialization"
)

const (
	DefaultMimetype = "application/octet-stream"
	DefaultCharset  = "utf-8"
)

// Server 提供 HTTP 服务
type Server struct {
	name       string
	version    string
	logs       *logs.Logs
	fs         fs.FS
	httpServer *http.Server
	vars       *sync.Map
	mimetypes  *serialization.Mimetypes
	encodings  *serialization.Encodings
	cache      cache.Cache
	uptime     time.Time
	serving    bool
	modules    []string // 保存着模块名称，用于检测是否存在重名
	routers    *Routers

	closed chan struct{} // 当 Close 延时关闭时，通过此事件确定 Close() 的退出时机。
	closes []func() error

	// service
	services  []*Service
	scheduled *scheduled.Server

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
		encodings:  serialization.NewEncodings(o.Logs.ERROR(), o.IgnoreEncodings...),
		cache:      o.Cache,
		uptime:     time.Now(),

		closed: make(chan struct{}, 1),
		closes: make([]func() error, 0, 10),

		// service
		services:  make([]*Service, 0, 100),
		scheduled: scheduled.NewServer(o.Location),

		// result
		resultMessages: make(map[string]*resultMessage, 20),
		resultBuilder:  o.ResultBuilder,

		// locale
		location:      o.Location,
		locale:        o.locale,
		tag:           o.Tag,
		localePrinter: o.locale.NewPrinter(o.Tag),
	}
	srv.routers = mux.NewRoutersOf(srv.call, nil)
	srv.httpServer.Handler = srv.routers

	return srv, nil
}

func (srv *Server) call(w http.ResponseWriter, r *http.Request, ps params.Params, f HandlerFunc) {
	ctx := srv.NewContext(w, r)
	if ctx == nil {
		return
	}

	ctx.params = ps
	if resp := f(ctx); resp != nil {
		if err := resp.Apply(ctx); err != nil {
			srv.logError(err)
		}
	}

	if err := ctx.destroy(); err != nil {
		srv.logError(err)
	}
}

func (srv *Server) logError(err error) {
	var msg string
	if ls, ok := err.(localeutil.LocaleStringer); ok {
		msg = ls.LocaleString(srv.LocalePrinter())
	} else {
		msg = err.Error()
	}
	srv.Logs().Error(msg)
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

// Serve 启动服务
//
// 会等待 Server.Close() 执行完之后，此函数才会返回，这一点与 Http.ListenAndServe 稍有不同。
// 一旦返回，整个 Server 对象将处于不可用状态。
func (srv *Server) Serve() (err error) {
	srv.runServices()

	// 在 Serve.defer 中关闭服务，而不是 Close。这样可以保证在所有的请求关闭之后执行。
	defer func() {
		srv.stopServices()

		sliceutil.Reverse(srv.closes)
		for _, f := range srv.closes {
			if err1 := f(); err1 != nil { // 出错不退出，继续其它操作。
				srv.Logs().Error(err1)
			}
		}

		// 在退出时还出错，直接 panic 问题也不大。
		if err2 := srv.Logs().Flush(); err2 != nil {
			panic(err2)
		}
	}()

	srv.serving = true

	cfg := srv.httpServer.TLSConfig
	if cfg != nil && (cfg.GetCertificate != nil || len(cfg.Certificates) > 0) {
		err = srv.httpServer.ListenAndServeTLS("", "")
	} else {
		err = srv.httpServer.ListenAndServe()
	}

	// 由 Server.Close() 主动触发的关闭事件，才需要等待其执行完成，
	// 其它错误直接返回，否则一些内部错误会永远卡在此处无法返回。
	if errors.Is(err, http.ErrServerClosed) {
		<-srv.closed
	}
	return err
}

// Close 触发关闭操作
//
// 需要等待 Server.Serve 返回才能证整个服务被关闭。
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
	if err := srv.httpServer.Shutdown(c); !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return nil
}

// Server 获取关联的 Server 实例
func (ctx *Context) Server() *Server { return ctx.server }

func (srv *Server) Mimetypes() *serialization.Mimetypes { return srv.mimetypes }

func (srv *Server) Encodings() *serialization.Encodings { return srv.encodings }

// Files 返回用于序列化文件内容的操作接口
func (srv *Server) Files() *serialization.Files { return srv.Locale().Files() }

// Locale 操作操作本地化文件的接口
func (srv *Server) Locale() *serialization.Locale { return srv.locale }

func (srv *Server) LocalePrinter() *message.Printer { return srv.localePrinter }

// Tag 返回默认的语言标签
func (srv *Server) Tag() language.Tag { return srv.tag }

// Serving 是否处于服务状态
func (srv *Server) Serving() bool { return srv.serving }

// OnClose 注册关闭服务时需要执行的函数
//
// NOTE: 按注册的相反顺序执行。
func (srv *Server) OnClose(f func() error) { srv.closes = append(srv.closes, f) }
