// SPDX-License-Identifier: MIT

package context

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware"
	"github.com/issue9/middleware/compress"
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

// Builder 定义了构建 Context 对象的一些通用数据选项
type Builder struct {
	Debug bool

	// 在调用 Builder.New 生成 Context
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

	logs        *logs.Logs
	interceptor func(ctx *Context)

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

// NewBuilder 返回 *Builder 实例
func NewBuilder(logs *logs.Logs, router *mux.Prefix, builder BuildResultFunc) *Builder {
	if builder == nil {
		builder = DefaultResultBuilder
	}

	return &Builder{
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
func (b *Builder) Logs() *logs.Logs {
	return b.logs
}

func (b *Builder) Handler() http.Handler {
	for url, dir := range b.Static {
		h := http.StripPrefix(url, http.FileServer(http.Dir(dir)))
		b.Router().Get(url+"{path}", h)
	}

	b.buildMiddlewares()

	return b.middlewares
}

// AddMiddlewares 设置全局的中间件，可多次调用。
func (b *Builder) AddMiddlewares(m middleware.Middleware) {
	b.middlewares.After(m)
}

// 通过配置文件加载相关的中间件
//
// 始终保持这些中间件在最后初始化。用户添加的中间件由 app.modules.After 添加。
func (b *Builder) buildMiddlewares() {
	b.middlewares.Before(func(h http.Handler) http.Handler { return b.errorHandlers.New(h) })

	// srv.errorhandlers.New 可能会输出大段内容。所以放在其之后。
	b.middlewares.Before(func(h http.Handler) http.Handler {
		return compress.New(h, &compress.Options{
			ErrorLog: b.Logs().ERROR(),
			Types:    []string{"*"},
			Funcs:    b.Compresses,
		})
	})

	// recovery
	b.middlewares.Before(func(h http.Handler) http.Handler {
		return recovery.New(h, b.errorHandlers.Recovery(b.Logs().ERROR()))
	})

	// NOTE: 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
	b.middlewares.Before(func(h http.Handler) http.Handler { return b.buildDebug(h) })
}

func (b *Builder) buildDebug(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case b.Debug && strings.HasPrefix(r.URL.Path, debugPprofPath):
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
		case b.Debug && strings.HasPrefix(r.URL.Path, debugVarsPath):
			expvar.Handler().ServeHTTP(w, r)
		default:
			h.ServeHTTP(w, r)
		}
	})
}
