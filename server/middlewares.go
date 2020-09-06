// SPDX-License-Identifier: MIT

package server

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/issue9/middleware"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/middleware/header"
	"github.com/issue9/middleware/host"
	"github.com/issue9/middleware/recovery"

	"github.com/issue9/web/internal/webconfig"
)

const (
	// 此变量理论上可以更改，但实际上，更改之后，所有的子页面都会不可用。
	debugPprofPath = "/debug/pprof/"

	// 此地址可以修改。
	debugVarsPath = "/debug/vars"
)

// AddMiddlewares 设置全局的中间件，可多次调用。
func (srv *Server) AddMiddlewares(m middleware.Middleware) {
	srv.middlewares.After(m)
}

// 通过配置文件加载相关的中间件。
//
// 始终保持这些中间件在最后初始化。用户添加的中间件由 app.modules.After 添加。
func (srv *Server) buildMiddlewares(conf *webconfig.WebConfig) {
	// domains
	if len(conf.AllowedDomains) > 0 {
		srv.middlewares.Before(func(h http.Handler) http.Handler {
			return host.New(h, conf.AllowedDomains...)
		})
	}

	// headers
	if len(conf.Headers) > 0 {
		srv.middlewares.Before(func(h http.Handler) http.Handler {
			return header.New(h, conf.Headers, nil)
		})
	}

	srv.middlewares.Before(func(h http.Handler) http.Handler {
		return srv.ErrorHandlers().New(h)
	})

	// compress
	// srv.errorhandlers.New 可能会输出大段内容。所以放在其之后。
	if conf.Compress != nil {
		srv.middlewares.Before(func(h http.Handler) http.Handler {
			return compress.New(h, &compress.Options{
				Funcs:    srv.compresses,
				Types:    conf.Compress,
				ErrorLog: srv.Logs().ERROR(),
			})
		})
	}

	// recovery
	srv.middlewares.Before(func(h http.Handler) http.Handler {
		return recovery.New(h, srv.ErrorHandlers().Recovery(srv.Logs().ERROR()))
	})

	// NOTE: 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
	if conf.Debug {
		srv.middlewares.Before(func(h http.Handler) http.Handler {
			return debug(h)
		})
	}
}

func debug(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, debugPprofPath):
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
		case strings.HasPrefix(r.URL.Path, debugVarsPath):
			expvar.Handler().ServeHTTP(w, r)
		default:
			h.ServeHTTP(w, r)
		}
	}) // end return http.HandlerFunc
}
