// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/cache/memory"
	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware/v3"
	"github.com/issue9/middleware/v3/compress"
	"github.com/issue9/middleware/v3/debugger"
	"github.com/issue9/middleware/v3/errorhandler"
	"github.com/issue9/mux/v3"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/content"
	"github.com/issue9/web/internal/dep"
	"github.com/issue9/web/result"
	"github.com/issue9/web/service"
)

type contextKey int

// ContextKeyServer 从 context.Value 中获取 *Server 实例的键名
//
// 在某些极端的情况下，用户可能需要用到 Server.Router().Mux().GetFunc()
// 等比较原始的接口去添加路由，此时无法像 Context.Server() 的方式获取
// Server 变量，便可通过 r.Context().Value(ContextKeyServer) 获取。
var ContextKeyServer contextKey = 0

// Options 初始化 Server 的参数
type Options struct {
	// 服务器的时区
	//
	// 默认值为 time.Local
	Location *time.Location

	// 当前使用的本地化组件
	//
	// 默认情况下会引用 golang.org/x/text/message.DefaultCatalog 对象。
	//
	// golang.org/x/text/message/catalog 提供了 NewBuilder 和 NewFromMap
	// 等方式构建 Catalog 接口实例。
	Catalog catalog.Catalog

	// 指定生成 Result 数据的方法
	//
	// 默认情况下指向  result.DefaultBuilder。
	ResultBuilder result.BuildFunc

	// 缓存系统
	//
	// 默认值为内存类型。
	Cache cache.Cache

	// 初始化 MUX 的参数
	DisableHead    bool
	DisableOptions bool
	SkipCleanPath  bool
	mux            *mux.Mux

	// 可以对 http.Server 的内容进行个性
	//
	// NOTE: 对 http.Server.Handler 的修改不会启作用，该值始终会指向 Server.middlewares
	HTTPServer func(*http.Server)
	httpServer *http.Server

	// 网站的根目录
	//
	// 可以带上域名：https://example.com/api；或是仅路径部分 /api；
	// 两者的区别在于 Router.URL 返回的内容，前者带域名部分，后者不带。
	Root string
	root *url.URL

	Config fs.FS
}

// Server 提供了用于构建 Context 对象的基本数据
type Server struct {
	name       string
	version    string
	logs       *logs.Logs
	httpServer *http.Server
	vars       map[interface{}]interface{}
	closed     chan struct{} // 当 shutdown 延时关闭时，通过此事件确定 Serve() 的返回时机。

	// middleware
	middlewares   *middleware.Manager
	compress      *compress.Compress
	errorHandlers *errorhandler.ErrorHandler
	debugger      *debugger.Debugger
	filters       []Filter

	// locale
	catalog  catalog.Catalog
	location *time.Location

	cache  cache.Cache
	router *Router
	uptime time.Time
	dep    *dep.Dep

	mimetypes *content.Mimetypes
	services  *service.Manager
	results   *result.Manager
}

func (o *Options) sanitize() (err error) {
	if o.Location == nil {
		o.Location = time.Local
	}

	if o.Catalog == nil {
		o.Catalog = message.DefaultCatalog
	}

	if o.ResultBuilder == nil {
		o.ResultBuilder = result.DefaultBuilder
	}

	if o.Cache == nil {
		o.Cache = memory.New(24 * time.Hour)
	}

	o.mux = mux.New(o.DisableOptions, o.DisableHead, o.SkipCleanPath, nil, nil)

	o.httpServer = &http.Server{}
	if o.HTTPServer != nil {
		o.HTTPServer(o.httpServer)
	}

	if o.root, err = url.Parse(o.Root); err != nil {
		return err
	}
	if o.httpServer.Addr == "" {
		if p := o.root.Port(); p != "" {
			o.httpServer.Addr = ":" + p
		} else if o.root.Scheme == "https" {
			o.httpServer.Addr = ":https"
		}
	}

	return nil
}

// NewServer 返回 *Server 实例
func NewServer(name, version string, logs *logs.Logs, o *Options) (*Server, error) {
	if err := o.sanitize(); err != nil {
		return nil, err
	}

	srv := &Server{
		name:       name,
		version:    version,
		logs:       logs,
		httpServer: o.httpServer,
		vars:       map[interface{}]interface{}{},
		closed:     make(chan struct{}, 1),

		middlewares:   middleware.NewManager(o.mux),
		compress:      compress.New(logs.ERROR(), "*"),
		errorHandlers: errorhandler.New(),
		debugger:      &debugger.Debugger{},

		catalog:  o.Catalog,
		location: o.Location,

		cache:  o.Cache,
		dep:    dep.New(logs.INFO()),
		uptime: time.Now(),

		mimetypes: content.NewMimetypes(),
		services:  service.NewManager(logs, o.Location),
		results:   result.NewManager(o.ResultBuilder),
	}
	srv.router = buildRouter(srv, o.mux, o.root)
	srv.httpServer.Handler = srv.middlewares

	if srv.httpServer.BaseContext == nil {
		srv.httpServer.BaseContext = func(n net.Listener) context.Context {
			return context.WithValue(context.Background(), ContextKeyServer, srv)
		}
	} else {
		ctx := srv.httpServer.BaseContext
		srv.httpServer.BaseContext = func(n net.Listener) context.Context {
			return context.WithValue(ctx(n), ContextKeyServer, srv)
		}
	}

	if err := srv.buildMiddlewares(); err != nil {
		return nil, err
	}

	return srv, nil
}

// GetServer 从请求中获取 *Server 实例
//
// r 必须得是由 Server 生成的，否则会 panic。
func GetServer(r *http.Request) *Server {
	v := r.Context().Value(ContextKeyServer)
	if v == nil {
		panic("无法从 http.Request.Context() 中获取 ContentKeyServer 对应的值")
	}

	return v.(*Server)
}

// AppName 应用的名称
func (srv *Server) AppName() string {
	return srv.name
}

// AppVersion 应用的版本
func (srv *Server) AppVersion() string {
	return srv.version
}

// Get 返回指定键名的值
func (srv *Server) Get(key interface{}) interface{} {
	return srv.vars[key]
}

// Set 保存指定键名的值
func (srv *Server) Set(key, val interface{}) {
	srv.vars[key] = val
}

// Location 指定服务器的时区信息
func (srv *Server) Location() *time.Location {
	return srv.location
}

// Logs 返回关联的 logs.Logs 实例
func (srv *Server) Logs() *logs.Logs {
	return srv.logs
}

// Cache 返回缓存的相关接口
func (srv *Server) Cache() cache.Cache {
	return srv.cache
}

// Uptime 当前服务的运行时间
func (srv *Server) Uptime() time.Time {
	return srv.uptime
}

// Now 返回当前时间
//
// 与 time.Now() 的区别在于 Now() 基于当前时区
func (srv *Server) Now() time.Time {
	return time.Now().In(srv.Location())
}

// ParseTime 分析基于当前时区的时间
func (srv *Server) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, srv.Location())
}

// Server 获取关联的 context.Server 实例
func (ctx *Context) Server() *Server {
	return ctx.server
}

// Mimetypes 返回内容编解码的管理接口
func (srv *Server) Mimetypes() *content.Mimetypes {
	return srv.mimetypes
}

// Services 返回服务内容的管理接口
func (srv *Server) Services() *service.Manager {
	return srv.services
}

// Serve 启动服务
//
// 会自动对模块进行初始化。
func (srv *Server) Serve() (err error) {
	if err = srv.initModules(srv.Logs().INFO()); err != nil {
		return err
	}

	srv.Services().Run()

	cfg := srv.httpServer.TLSConfig
	if cfg != nil && (cfg.GetCertificate != nil || len(cfg.Certificates) > 0) {
		err = srv.httpServer.ListenAndServeTLS("", "")
	}
	err = srv.httpServer.ListenAndServe()

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
