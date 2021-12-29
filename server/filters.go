// SPDX-License-Identifier: MIT

package server

import (
	"net/http"

	"github.com/issue9/mux/v5/middleware"
)

// Filter 适用于 Context 的中间件
//
// Filter 和 github.com/issue9/middleware.Func 本质上没有任何区别，
// middleware.Func 更加的通用，可以复用市面上的大部分中间件，
// Filter 则更加灵活一些，适合针对当前框架的中间件。
//
// 如果想要使用 middleware.Func，可以调用 Server.MuxGroups().Middlewares() 方法。
type Filter func(HandlerFunc) HandlerFunc

func (srv *Server) Middlewares() *middleware.Middlewares {
	return srv.group.Middlewares()
}

// AcceptFilter 提供限定 accept 的中间件
func AcceptFilter(ct ...string) Filter {
	return func(next HandlerFunc) HandlerFunc {
		return Accept(next, ct...)
	}
}

// Accept 提供限定 accept 的中间件
func Accept(next HandlerFunc, ct ...string) HandlerFunc {
	return func(ctx *Context) Responser {
		for _, c := range ct {
			if c == ctx.OutputMimetypeName {
				return next(ctx)
			}
		}
		return Status(http.StatusNotAcceptable)
	}
}
