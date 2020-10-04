// SPDX-License-Identifier: MIT

package context

import (
	"net/http"

	"github.com/issue9/middleware/v2"
	"github.com/issue9/middleware/v2/errorhandler"
)

// Filter 针对 Context 的中间件
//
// Filter 和 github.com/issue9/middleware.Middleware 本质上没有任何区别，
// 都是作用于 http.Handler 上的中间件，只因参数不同，且两者不能交替出现，
// 派生出两套类型。
//
// 保证针对 middleware.Middleware 的 AddMiddlewares 方法，
// 可以最大限度地利用现有的通用中间件，而 AddFilter
// 方便用户编写针对 Context 的中间件，且 Context 提供了
// http.Handler 不存在的功能。
type Filter func(HandlerFunc) HandlerFunc

// FilterHandler 将过滤器应用于处理函数 next
func FilterHandler(next HandlerFunc, filter ...Filter) HandlerFunc {
	if l := len(filter); l > 0 {
		for i := l - 1; i >= 0; i-- {
			next = filter[i](next)
		}
	}
	return next
}

// AddFilters 添加过滤器
func (srv *Server) AddFilters(filter ...Filter) {
	srv.filters = append(srv.filters, filter...)
}

// 始终保持这些中间件在最后初始化。用户添加的中间件由 Server.AddMiddlewares 添加。
func (srv *Server) buildMiddlewares() {
	rf := srv.errorHandlers.Recovery(srv.logs.ERROR())
	srv.AddMiddlewares(
		srv.debugger.Middleware, // 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
		rf.Middleware,
		srv.compress.Middleware, // srv.errorhandlers.New 可能会输出大段内容。所以放在其之前。
		srv.errorHandlers.Middleware,
	)
}

// SetErrorHandle 设置指定状态码页面的处理函数
//
// 如果状态码已经存在处理函数，则修改，否则就添加。
// 仅对状态码 >= 400 的有效果。
// 如果 status 为零表示所有未设置的状态码都采用该函数处理。
func (srv *Server) SetErrorHandle(h errorhandler.HandleFunc, status ...int) {
	srv.errorHandlers.Set(h, status...)
}

// AddMiddlewares 添加中间件
func (srv *Server) AddMiddlewares(middleware ...middleware.Middleware) {
	for _, m := range middleware {
		srv.middlewares.After(m)
	}
}

// SetDebugger 设置调试地址
func (srv *Server) SetDebugger(pprof, vars string) {
	srv.debugger.Pprof = pprof
	srv.debugger.Vars = vars
}

// Handler 将当前服务转换为 http.Handler 接口对象
func (srv *Server) Handler() http.Handler {
	return srv.middlewares
}
