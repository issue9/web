// SPDX-License-Identifier: MIT

package context

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/middleware/header"
	"github.com/issue9/middleware/host"
	"github.com/issue9/middleware/recovery"
	"github.com/issue9/middleware/recovery/errorhandler"
	"github.com/issue9/mux/v2"
)

const (
	// 此变量理论上可以更改，但实际上，更改之后，所有的子页面都会不可用。
	debugPprofPath = "/debug/pprof/"

	// 此地址可以修改。
	debugVarsPath = "/debug/vars"
)

// Server 定义了构建 Context 对象的一些通用数据选项
//
// 所有公开的变量，均可为零值，在调用 Handler() 转换之前，
// 这些值可以随意设置，但是在调用 Handler() 之后再修改，
// 则相关的修改未必是有效的。
type Server struct {
	// 调试模式
	//
	// 会额外提供 /debug/pprof/* 和 /debug/vars 用于输出调试信息。
	Debug bool

	// 在调用 Server.newContext 生成 Context
	// 之前可能通过此方法对其进行一次统一的修改，不需要则为 nil。
	Interceptor func(*Context)

	// 对内容压缩的配置项，如果为空，则不会作压缩处理。
	Compresses map[string]compress.WriterFunc

	// Static 静态内容，键名为 URL 路径，键值为文件地址
	//
	// 比如在 Domain 和 Root 的值分别为 example.com 和 blog 时，
	// 将 Static 的值设置为 /admin ==> ~/data/assets/admin
	// 表示将 example.com/blog/admin/* 解析到 ~/data/assets/admin 目录之下。
	Static map[string]string

	// AllowedDomains 限定访问域名
	//
	// 若指定了此值，则只有此列表中指定的域名可以访问当前网页。
	// 诸如 IP 和其它域名的指向将不再启作用。
	//
	// 可以指定泛域名，比如 *.example.com
	AllowedDomains []string

	// Location 指定服务器的时区信息
	//
	// 如果未指定，则会采用 time.Local 作为默认值。
	// 在构建 Context 对象时，该时区信息也会分配给 Context，
	// 如果每个 Context 对象需要不同的值，可以在 Interceptor 中进行修改。
	Location *time.Location

	// 额外输出给客户端的报头
	//
	// 这些报头在用户添加的中间件的外层，用户可以通过添加中件间覆盖这些报头。
	Headers map[string]string

	logs        *logs.Logs
	interceptor func(ctx *Context)
	uptime      time.Time

	// routes
	middlewares   *middleware.Manager
	errorHandlers *errorhandler.ErrorHandler
	router        *mux.Prefix

	// result
	resultBuilder BuildResultFunc
	messages      map[int]*resultMessage

	// mimetype
	marshals   []*marshaler
	unmarshals []*unmarshaler
}

// NewServer 返回 *Server 实例
func NewServer(logs *logs.Logs, router *mux.Prefix, builder BuildResultFunc) *Server {
	if builder == nil {
		builder = DefaultResultBuilder
	}

	return &Server{
		logs: logs,

		middlewares:   middleware.NewManager(router.Mux()),
		errorHandlers: errorhandler.New(),
		router:        router,

		resultBuilder: builder,
		messages:      make(map[int]*resultMessage, 20),

		marshals:   make([]*marshaler, 0, 10),
		unmarshals: make([]*unmarshaler, 0, 10),
	}
}

// Logs 返回关联的 logs.Logs 实例
func (srv *Server) Logs() *logs.Logs {
	return srv.logs
}

// Handler 将当前服务转换为 http.Handler 接口对象
func (srv *Server) Handler() http.Handler {
	for url, dir := range srv.Static {
		h := http.StripPrefix(url, http.FileServer(http.Dir(dir)))
		srv.Router().Get(url+"{path}", h)
	}

	srv.uptime = time.Now()

	if srv.Location == nil {
		srv.Location = time.Local
	}

	srv.buildMiddlewares()

	return srv.middlewares
}

// AddMiddlewares 设置全局的中间件，可多次调用
func (srv *Server) AddMiddlewares(m middleware.Middleware) {
	srv.middlewares.After(m)
}

// 通过配置文件加载相关的中间件
//
// 始终保持这些中间件在最后初始化。用户添加的中间件由 app.modules.After 添加。
func (srv *Server) buildMiddlewares() {
	srv.middlewares.Before(func(h http.Handler) http.Handler {
		return header.New(h, srv.Headers, nil)
	})

	if len(srv.AllowedDomains) > 0 {
		srv.middlewares.Before(func(h http.Handler) http.Handler {
			return host.New(h, srv.AllowedDomains...)
		})
	}

	srv.middlewares.Before(func(h http.Handler) http.Handler { return srv.errorHandlers.New(h) })

	// srv.errorhandlers.New 可能会输出大段内容。所以放在其之后。
	srv.middlewares.Before(func(h http.Handler) http.Handler {
		return compress.New(h, &compress.Options{
			ErrorLog: srv.Logs().ERROR(),
			Types:    []string{"*"},
			Funcs:    srv.Compresses,
		})
	})

	// recovery
	srv.middlewares.Before(func(h http.Handler) http.Handler {
		return recovery.New(h, srv.errorHandlers.Recovery(srv.Logs().ERROR()))
	})

	// NOTE: 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
	srv.middlewares.Before(func(h http.Handler) http.Handler { return srv.buildDebug(h) })
}

func (srv *Server) buildDebug(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case srv.Debug && strings.HasPrefix(r.URL.Path, debugPprofPath):
			path := r.URL.Path[len(debugPprofPath):]
			switch path {
			case "cmdline":
				pprof.Cmdline(w, r)
			case "profile":
				pprof.Profile(w, r)
			case "symbol":
				pprof.Symbol(w, r)
			case "trace":
				pprof.Trace(w, r)
			default:
				pprof.Index(w, r)
			}
		case srv.Debug && strings.HasPrefix(r.URL.Path, debugVarsPath):
			expvar.Handler().ServeHTTP(w, r)
		default:
			h.ServeHTTP(w, r)
		}
	})
}

// Uptime 当前服务的运行时间
func (srv *Server) Uptime() time.Time {
	return srv.uptime
}
