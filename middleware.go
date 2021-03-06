// SPDX-License-Identifier: MIT

package web

import (
	"net/http"

	"github.com/issue9/middleware/v3"
	"github.com/issue9/middleware/v3/compress"
	"github.com/issue9/middleware/v3/errorhandler"
	"github.com/issue9/middleware/v3/recovery"
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

// AddFilters 添加过滤器
//
// 按给定参数的顺序反向依次调用。
func (m *Module) AddFilters(filter ...Filter) *Module {
	m.filters = append(m.filters, filter...)
	return m
}

func (srv *Server) buildMiddlewares() error {
	if err := srv.compress.AddAlgorithm("deflate", compress.NewDeflate); err != nil {
		return err
	}

	if err := srv.compress.AddAlgorithm("gzip", compress.NewGzip); err != nil {
		return err
	}

	if err := srv.compress.AddAlgorithm("br", compress.NewBrotli); err != nil {
		return err
	}

	srv.AddMiddlewares(
		srv.debugger.Middleware, // 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
		srv.recoveryMiddleware,
		srv.compress.Middleware, // srv.errorhandlers.New 可能会输出大段内容。所以放在其之前。
		srv.errorHandlers.Middleware,
	)

	return nil
}

func (srv *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.recoverFunc.Middleware(next).ServeHTTP(w, r)
	})
}

// SetRecovery 设置在 panic 时的处理函数
//
// 默认情况下，会向用户输出 500 的错误信息。
func (srv *Server) SetRecovery(f recovery.RecoverFunc) {
	srv.recoverFunc = f
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
func (srv *Server) SetDebugger(pprof, vars string) (err error) {
	if pprof != "" {
		if pprof, err = srv.Router().Path(pprof, nil); err != nil {
			return err
		}
	}

	if vars != "" {
		if vars, err = srv.Router().Path(vars, nil); err != nil {
			return err
		}
	}

	srv.debugger.Pprof = pprof
	srv.debugger.Vars = vars

	return nil
}

// SetCompressAlgorithm 设置压缩的算法
//
// 默认情况下，支持 gzip、deflate 和 br 三种。
// 如果 w 为 nil，表示删除对该算法的支持。
func (srv *Server) SetCompressAlgorithm(name string, w compress.WriterFunc) {
	srv.compress.SetAlgorithm(name, w)
}

// AddCompressTypes 指定哪些内容可以进行压缩传输
//
// 默认情况下是所有内容都将进行压缩传输，
// * 表示所有；
// text/* 表示以 text/ 开头的类型；
// text/plain 表示具体类型 text/plain；
func (srv *Server) AddCompressTypes(types ...string) {
	srv.compress.AddType(types...)
}

// DeleteCompressTypes 删除指定内容的压缩支持
//
// 仅用于删除通过 AddType 添加的内容。
//
// NOTE: 如果指定 * 之后的所有内容都将不支持。
func (srv *Server) DeleteCompressTypes(types ...string) {
	srv.compress.DeleteType(types...)
}
