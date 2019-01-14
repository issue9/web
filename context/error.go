// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"fmt"

	"github.com/issue9/middleware/recovery/errorhandler"
)

// Critical 输出一条日志到 CRITICAL 日志通道，并向用户输出一个指定状态码的页面。
//
// 若是输出日志的过程中出错，则 panic
// 若没有错误信息，则仅向客户端输出一条状态码信息。
func (ctx *Context) Critical(status int, v ...interface{}) {
	if len(v) > 0 {
		ctx.App.Logs().CRITICAL().Output(2, fmt.Sprint(v...))
	}

	ctx.App.ErrorHandlers().Render(ctx.Response, status)
}

// Error 输出一条日志到 ERROR 日志通道，并向用户输出一个指定状态码的页面。
//
// 若是输出日志的过程中出错，则 panic
// 若没有错误信息，则仅向客户端输出一条状态码信息。
func (ctx *Context) Error(status int, v ...interface{}) {
	if len(v) > 0 {
		ctx.App.Logs().ERROR().Output(2, fmt.Sprint(v...))
	}

	ctx.App.ErrorHandlers().Render(ctx.Response, status)
}

// Criticalf 输出一条日志到 CRITICAL 日志通道，并向用户输出一个指定状态码的页面。
//
// 若是输出日志的过程中出错，则 panic
// 若没有错误信息，则仅向客户端输出一条状态码信息。
func (ctx *Context) Criticalf(status int, format string, v ...interface{}) {
	if len(v) > 0 {
		ctx.App.Logs().CRITICAL().Output(2, fmt.Sprintf(format, v...))
	}

	ctx.App.ErrorHandlers().Render(ctx.Response, status)
}

// Errorf 输出一条日志到 ERROR 日志通道，并向用户输出一个指定状态码的页面。
//
// 若是输出日志的过程中出错，则 panic
// 若没有错误信息，则仅向客户端输出一条状态码信息。
func (ctx *Context) Errorf(status int, format string, v ...interface{}) {
	if len(v) > 0 {
		ctx.App.Logs().ERROR().Output(2, fmt.Sprintf(format, v...))
	}

	ctx.App.ErrorHandlers().Render(ctx.Response, status)
}

// Exit 以指定的状态码退出当前协程
func (ctx *Context) Exit(status int) {
	Exit(status)
}

// Exit 以指定的状态码退出当前协程
//
// status 表示输出的状态码，如果为 0，则不会作任何状态码输出。
//
// Exit 最终是以 panic 的形式退出，所以如果你的代码里截获了 panic，
// 那么 Exit 并不能达到退出当前请求的操作。
//
// 与 Error 的不同在于：
// Error 不会主动退出当前协程，而 Exit 则会触发 panic，退出当前协程。
func Exit(status int) {
	errorhandler.Exit(status)
}
