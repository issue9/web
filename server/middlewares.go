// SPDX-License-Identifier: MIT

package server

import (
	"net/http"

	"github.com/issue9/middleware"
	"github.com/issue9/middleware/header"
	"github.com/issue9/middleware/host"

	"github.com/issue9/web/internal/webconfig"
)

// AddMiddlewares 设置全局的中间件，可多次调用。
func (srv *Server) AddMiddlewares(m middleware.Middleware) {
	srv.Builder().AddMiddlewares(m)
}

// 通过配置文件加载相关的中间件。
//
// 始终保持这些中间件在最后初始化。用户添加的中间件由 app.modules.After 添加。
func (srv *Server) buildMiddlewares(conf *webconfig.WebConfig) {
	// headers
	if len(conf.Headers) > 0 {
		srv.Builder().AddMiddlewares(func(h http.Handler) http.Handler {
			return header.New(h, conf.Headers, nil)
		})
	}

	// domains
	if len(conf.AllowedDomains) > 0 {
		srv.Builder().AddMiddlewares(func(h http.Handler) http.Handler {
			return host.New(h, conf.AllowedDomains...)
		})
	}
}
