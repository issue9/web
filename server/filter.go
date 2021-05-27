// SPDX-License-Identifier: MIT

package server

import (
	"github.com/issue9/middleware/v4/compress"
	"github.com/issue9/middleware/v4/errorhandler"
)

// Filter 针对 Context 的中间件
//
// Filter 和 github.com/issue9/mux.MiddlewareFunc 本质上没有任何区别，
// mux.MiddlewareFunc 更加的通用，可以复用市面上的大部分中间件，
// Filter 则更加灵活一些，可以针对模块或是某一个路由。
//
// 如果想要使用 mux.MiddlewareFunc，可以调用 Server.Mux().AddMiddleware 方法。
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

// SetErrorHandle 设置指定状态码页面的处理函数
//
// 如果状态码已经存在处理函数，则修改，否则就添加。
func (srv *Server) SetErrorHandle(h errorhandler.HandleFunc, status ...int) {
	srv.errorHandlers.Set(h, status...)
}

// SetCompressAlgorithm 设置压缩的算法
//
// 默认情况下，支持 gzip、deflate 和 br 三种。
// 如果 w 为 nil，表示删除对该算法的支持。
func (srv *Server) SetCompressAlgorithm(name string, w compress.WriterFunc) error {
	return srv.compress.SetAlgorithm(name, w)
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
