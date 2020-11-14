// SPDX-License-Identifier: MIT

package context

import (
	"net/http"
	"net/url"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/cache/memory"
	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware/v2"
	"github.com/issue9/middleware/v2/compress"
	"github.com/issue9/middleware/v2/debugger"
	"github.com/issue9/middleware/v2/errorhandler"
	"github.com/issue9/mux/v3"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/context/contentype"
	"github.com/issue9/web/context/service"
)

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
	// 默认情况下指向  DefaultResultBuilder。
	ResultBuilder BuildResultFunc

	// 缓存系统
	//
	// 默认值为内存类型。
	Cache cache.Cache

	// 初始化 MUX 的参数
	DisableHead    bool
	DisableOptions bool
	SkipCleanPath  bool

	// 网站的根目录
	Root string
	root *url.URL
}

// Server 提供了用于构建 Context 对象的基本数据
type Server struct {
	// 保存 Context 在存续期间的可复用变量
	//
	// 这是比 context.Value 更经济的传递变量方式。
	//
	// 如果仅需要在单次请求中传递参数，可直接使用 Context.Vars。
	Vars map[interface{}]interface{}

	logs *logs.Logs

	// middleware
	middlewares   *middleware.Manager
	compress      *compress.Compress
	errorHandlers *errorhandler.ErrorHandler
	debugger      *debugger.Debugger
	filters       []Filter

	catalog       catalog.Catalog
	resultBuilder BuildResultFunc
	location      *time.Location
	cache         cache.Cache

	router    *Router
	uptime    time.Time
	mimetypes *contentype.Mimetypes
	services  *service.Manager

	// result
	messages map[int]*resultMessage
}

func (o *Options) sanitize() (err error) {
	if o.Location == nil {
		o.Location = time.Local
	}

	if o.Catalog == nil {
		o.Catalog = message.DefaultCatalog
	}

	if o.ResultBuilder == nil {
		o.ResultBuilder = DefaultResultBuilder
	}

	if o.Cache == nil {
		o.Cache = memory.New(24 * time.Hour)
	}

	o.root, err = url.Parse(o.Root)
	return err
}

// NewServer 返回 *Server 实例
func NewServer(logs *logs.Logs, o *Options) (*Server, error) {
	if err := o.sanitize(); err != nil {
		return nil, err
	}

	mux := mux.New(o.DisableOptions, o.DisableHead, o.SkipCleanPath, nil, nil)

	srv := &Server{
		logs: logs,

		Vars: map[interface{}]interface{}{},

		middlewares: middleware.NewManager(mux),
		compress: compress.New(logs.ERROR(), map[string]compress.WriterFunc{
			"gzip":    compress.NewGzip,
			"deflate": compress.NewDeflate,
			"br":      compress.NewBrotli,
		}, "*"),
		errorHandlers: errorhandler.New(),
		debugger:      &debugger.Debugger{},

		catalog:       o.Catalog,
		resultBuilder: o.ResultBuilder,
		location:      o.Location,
		cache:         o.Cache,

		uptime:    time.Now(),
		mimetypes: contentype.NewMimetypes(),
		services:  service.NewManager(o.Location, logs),

		messages: make(map[int]*resultMessage, 20),
	}
	srv.router = buildRouter(srv, mux, o.root)

	srv.buildMiddlewares()

	return srv, nil
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
func (srv *Server) Mimetypes() *contentype.Mimetypes {
	return srv.mimetypes
}

// Services 返回服务内容的管理接口
func (srv *Server) Services() *service.Manager {
	return srv.services
}

// Close 关闭服务
func (srv *Server) Close() {
	srv.Services().Stop()
}

// Handler 将当前服务转换为 http.Handler 接口对象
func (srv *Server) Handler() http.Handler {
	return srv.middlewares
}

// Serve 启动服务
//
// httpServer.Handler 会被 srv 的相关内容替换
//
// 根据是否有配置 httpServer.TLSConfig.GetCertificate 或是 httpServer.TLSConfig.Certificates
// 决定是调用 ListenAndServeTLS 还是 ListenAndServe。
func (srv *Server) Serve(httpServer *http.Server) error {
	httpServer.Handler = srv.middlewares

	srv.Services().Run()

	cfg := httpServer.TLSConfig
	if cfg.GetCertificate != nil || len(cfg.Certificates) > 0 {
		return httpServer.ListenAndServeTLS("", "")
	}
	return httpServer.ListenAndServe()
}
