// SPDX-License-Identifier: MIT

package server

import (
	"fmt"

	"github.com/issue9/middleware/v3/errorhandler"
)

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
	ExitContext(status)
}

// ExitContext 以指定的状态码退出当前协程
//
// status 表示输出的状态码，如果为 0，则不会作任何状态码输出。
//
// ExitContext 最终是以 panic 的形式退出，所以如果你的代码里截获了 panic，
// 那么 ExitContext 并不能达到退出当前请求的操作。
//
// 与 Contet.Error 的不同在于：
// Contet.Error 不会主动退出当前协程，而 ExitContext 则会触发 panic，退出当前协程。
func ExitContext(status int) {
	errorhandler.Exit(status)
}
