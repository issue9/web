// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"net/http"

	"github.com/issue9/logs"
	"github.com/issue9/web/errors"
)

// Critical 输出一条日志到 CRITICAL 日志通道，
// 并向用户输出一个指定状态码的页面。
// 若是输出日志的过程中出错，则 panic
func (ctx *Context) Critical(status int, v ...interface{}) {
	logs.CRITICAL().Output(2, errors.TraceStack(2, v...))

	ctx.RenderStatus(status)
}

// Error 输出一条日志到 ERROR 日志通道，
// 并向用户输出一个指定状态码的页面。
// 若是输出日志的过程中出错，则 panic
func (ctx *Context) Error(status int, v ...interface{}) {
	logs.ERROR().Output(2, errors.TraceStack(2, v...))

	ctx.RenderStatus(status)
}

// Panic 以指定的状态码抛出异常
//
// 与 Error 的不同在于：
// Error 不会主动退出当前协程，而 Panic 则会触发 panic，退出当前协程。
func (ctx *Context) Panic(status int) {
	Panic(status)
}

// Critical 输出一条日志到 CRITICAL 日志通道，
// 并向用户输出一个指定状态码的页面。
// 若是输出日志的过程中出错，则 panic
func Critical(w http.ResponseWriter, status int, v ...interface{}) {
	logs.CRITICAL().Output(2, errors.TraceStack(2, v...))

	RenderStatus(w, status)
}

// Error 输出一条日志到 ERROR 日志通道，
// 并向用户输出一个指定状态码的页面。
// 若是输出日志的过程中出错，则 panic
func Error(w http.ResponseWriter, status int, v ...interface{}) {
	logs.ERROR().Output(2, errors.TraceStack(2, v...))

	RenderStatus(w, status)
}

// Panic 以指定的状态码抛出异常
//
// 与 Error 的不同在于：
// Error 不会主动退出当前协程，而 Panic 则会触发 panic，退出当前协程。
func Panic(status int) {
	panic(errors.HTTP(status))
}
