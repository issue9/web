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

	"github.com/issue9/web/content"
	"github.com/issue9/web/dep"
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
	closed     chan struct{} // 当 shutdown 延时关闭时，通过此事件确定 Serve() 的返回时机。

	// middleware
	groups        *group.Groups
	compress      *compress.Compress
	errorHandlers *errorhandler.ErrorHandler

	// locale
	location *time.Location

	cache    cache.Cache
	uptime   time.Time
	dep      *dep.Dep
	content  *content.Content
	services *service.Manager
	events   map[string]events.Eventer
}

// New 返回 *Server 实例
//
// name, version 表示服务的名称和版本号；
// o 指定了初始化 Server 一些非必要参数。在传递给 New 之后，再对其值进行改变，是无效的。
func New(name, version string, logs *logs.Logs, o *Options) (*Server, error) {
	o, err := o.sanitize()
	if err != nil {
		return nil, err
	}

	srv := &Server{
		name:       name,
		version:    version,
		logs:       logs,
		fs:         o.FS,
		httpServer: o.httpServer,
		vars:       &sync.Map{},
		closed:     make(chan struct{}, 1),

		groups:        o.groups,
		compress:      compress.Classic(logs.ERROR(), o.IgnoreCompressTypes...),
		errorHandlers: errorhandler.New(),

		location: o.Location,

		cache:    o.Cache,
		dep:      dep.New(),
		uptime:   time.Now(),
		content:  content.New(o.ResultBuilder),
		services: service.NewManager(logs, o.Location),
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

// Server 获取关联的 context.Server 实例
func (ctx *Context) Server() *Server { return ctx.server }

// Content 返回内容编解码的管理接口
func (srv *Server) Content() *content.Content { return srv.content }

// Services 返回服务内容的管理接口
func (srv *Server) Services() *service.Manager { return srv.services }

// Serve 启动服务
//
// 在运行之前，请确保已经调用 InitModules。
func (srv *Server) Serve() (err error) {
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

	srv.Services().Stop()

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

// Now 返回当前时间
//
// 与 time.Now() 的区别在于 Now() 基于当前时区
func (ctx *Context) Now() time.Time { return time.Now().In(ctx.Location) }

// ParseTime 分析基于当前时区的时间
func (ctx *Context) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, ctx.Location)
}
