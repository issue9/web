// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"fmt"

	"github.com/issue9/logs"
	"github.com/issue9/web/contentype"
)

func (ctx *Context) logMessage(v []interface{}) string {
	if ctx.r != nil {
		v = append(v, "@", ctx.r.URL.String())
	}

	return fmt.Sprintln(v...)
}

func (ctx *Context) logMessagef(format string, v []interface{}) string {
	if ctx.r != nil {
		format = format + "@" + ctx.r.URL.String()
	}

	return fmt.Sprintf(format, v...)
}

// Critical 相当于调用了 logs.Critical，外加一些调用者的详细信息
//
// 出错，一般也意味着当前协程的结束，所以会返回一个 Renderer
// 接口，方便向当前客户端输出相关的错误信息。
func (ctx *Context) Critical(v ...interface{}) contentype.Renderer {
	logs.CRITICAL().Output(2, ctx.logMessage(v))
	return ctx.ct
}

// Criticalf 相当于调用了 logs.Criticalf，外加一些调用者的详细信息
//
// 出错，一般也意味着当前协程的结束，所以会返回一个 Renderer
// 接口，方便向当前客户端输出相关的错误信息。
func (ctx *Context) Criticalf(format string, v ...interface{}) contentype.Renderer {
	logs.CRITICAL().Output(2, ctx.logMessagef(format, v))
	return ctx.ct
}

// Error 相当于调用了 logs.Error，外加一些调用者的详细信息
//
// 出错，一般也意味着当前协程的结束，所以会返回一个 Renderer
// 接口，方便向当前客户端输出相关的错误信息。
func (ctx *Context) Error(v ...interface{}) contentype.Renderer {
	logs.ERROR().Output(2, ctx.logMessage(v))
	return ctx.ct
}

// Errorf 相当于调用了 logs.Errorf，外加一些调用者的详细信息
//
// 出错，一般也意味着当前协程的结束，所以会返回一个 Renderer
// 接口，方便向当前客户端输出相关的错误信息。
func (ctx *Context) Errorf(format string, v ...interface{}) contentype.Renderer {
	logs.ERROR().Output(2, ctx.logMessagef(format, v))
	return ctx.ct
}

// Debug 相当于调用了 logs.Debug，外加一些调用者的详细信息
func (ctx *Context) Debug(v ...interface{}) {
	logs.DEBUG().Output(2, ctx.logMessage(v))
}

// Debugf 相当于调用了 logs.Debugf，外加一些调用者的详细信息
func (ctx *Context) Debugf(format string, v ...interface{}) {
	logs.DEBUG().Output(2, ctx.logMessagef(format, v))
}

// Trace 相当于调用了 logs.Trace，外加一些调用者的详细信息
func (ctx *Context) Trace(v ...interface{}) {
	logs.TRACE().Output(2, ctx.logMessage(v))
}

// Tracef 相当于调用了 logs.Tracef，外加一些调用者的详细信息
func (ctx *Context) Tracef(format string, v ...interface{}) {
	logs.TRACE().Output(2, ctx.logMessagef(format, v))
}

// Warn 相当于调用了 logs.Warn，外加一些调用者的详细信息
func (ctx *Context) Warn(v ...interface{}) {
	logs.WARN().Output(2, ctx.logMessage(v))
}

// Warnf 相当于调用了 logs.Warnf，外加一些调用者的详细信息
func (ctx *Context) Warnf(format string, v ...interface{}) {
	logs.WARN().Output(2, ctx.logMessagef(format, v))
}

// Info 相当于调用了 logs.Info，外加一些调用者的详细信息
func (ctx *Context) Info(v ...interface{}) {
	logs.INFO().Output(2, ctx.logMessage(v))
}

// Infof 相当于调用了 logs.Infof，外加一些调用者的详细信息
func (ctx *Context) Infof(format string, v ...interface{}) {
	logs.INFO().Output(2, ctx.logMessagef(format, v))
}
