// SPDX-License-Identifier: MIT

package server

import "fmt"

// Critical 输出日志到 CRITICAL 通道并向用户输出指定状态码的页面
//
// 若是输出日志的过程中出错，则 panic
// 若没有错误信息，则仅向客户端输出一条状态码信息。
//
// 输出的内容，会根据 status 的值，在 errorhandler 中查找相应的响应内容，
// 即使该值小于 400。
func (ctx *Context) Critical(status int, v ...interface{}) {
	if len(v) > 0 {
		ctx.server.Logs().CRITICAL().Output(2, fmt.Sprint(v...))
	}
	ctx.Exit(status)
}

// Error 输出日志到 ERROR 通道并向用户输出指定状态码的页面
//
// 若是输出日志的过程中出错，则 panic
// 若没有错误信息，则仅向客户端输出一条状态码信息。
//
// 输出的内容，会根据 status 的值，在 errorhandler 中查找相应的响应内容，
// 即使该值小于 400。
func (ctx *Context) Error(status int, v ...interface{}) {
	if len(v) > 0 {
		ctx.server.Logs().ERROR().Output(2, fmt.Sprint(v...))
	}
	ctx.Exit(status)
}

// Criticalf 输出日志到 CRITICAL 通道并向用户输出指定状态码的页面
//
// 若是输出日志的过程中出错，则 panic
// 若没有错误信息，则仅向客户端输出一条状态码信息。
//
// 输出的内容，会根据 status 的值，在 errorhandler 中查找相应的响应内容，
// 即使该值小于 400。
func (ctx *Context) Criticalf(status int, format string, v ...interface{}) {
	if len(v) > 0 {
		ctx.server.Logs().CRITICAL().Output(2, fmt.Sprintf(format, v...))
	}
	ctx.Exit(status)
}

// Errorf 输出日志到 ERROR 通道并向用户输出指定状态码的页面
//
// 若是输出日志的过程中出错，则 panic
// 若没有错误信息，则仅向客户端输出一条状态码信息。
//
// 输出的内容，会根据 status 的值，在 errorhandler 中查找相应的响应内容，
// 即使该值小于 400。
func (ctx *Context) Errorf(status int, format string, v ...interface{}) {
	if len(v) > 0 {
		ctx.server.Logs().ERROR().Output(2, fmt.Sprintf(format, v...))
	}
	ctx.Exit(status)
}

// Exit 以指定的状态码退出当前协程
func (ctx *Context) Exit(status int) {
	ctx.Server().errorHandlers.Exit(ctx.Response, status)
}
