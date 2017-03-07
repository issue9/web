// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"fmt"
	"net/http"

	"github.com/issue9/logs"
	"github.com/issue9/web/contentype"
)

func logMessage(r *http.Request, v []interface{}) string {
	if r != nil {
		v = append(v, "@", r.URL)
	}

	return fmt.Sprintln(v...)
}

func logMessagef(r *http.Request, format string, v []interface{}) string {
	if r != nil {
		format = format + "@" + r.URL.String()
	}

	return fmt.Sprintf(format, v...)
}

// Critical 相当于调用了 logs.Critical，外加一些调用者的详细信息
//
// 出错，一般也意味着当前协程的结束，所以会返回一个 Renderer
// 接口，方便向当前客户端输出相关的错误信息。
func Critical(r *http.Request, v ...interface{}) contentype.Renderer {
	logs.CRITICAL().Output(2, logMessage(r, v))
	return ContentType()
}

// Criticalf 相当于调用了 logs.Criticalf，外加一些调用者的详细信息
//
// 出错，一般也意味着当前协程的结束，所以会返回一个 Renderer
// 接口，方便向当前客户端输出相关的错误信息。
func Criticalf(r *http.Request, format string, v ...interface{}) contentype.Renderer {
	logs.CRITICAL().Output(2, logMessagef(r, format, v))
	return ContentType()
}

// Error 相当于调用了 logs.Error，外加一些调用者的详细信息
//
// 出错，一般也意味着当前协程的结束，所以会返回一个 Renderer
// 接口，方便向当前客户端输出相关的错误信息。
func Error(r *http.Request, v ...interface{}) contentype.Renderer {
	logs.ERROR().Output(2, logMessage(r, v))
	return ContentType()
}

// Errorf 相当于调用了 logs.Errorf，外加一些调用者的详细信息
//
// 出错，一般也意味着当前协程的结束，所以会返回一个 Renderer
// 接口，方便向当前客户端输出相关的错误信息。
func Errorf(r *http.Request, format string, v ...interface{}) contentype.Renderer {
	logs.ERROR().Output(2, logMessagef(r, format, v))
	return ContentType()
}

// Debug 相当于调用了 logs.Debug，外加一些调用者的详细信息
func Debug(r *http.Request, v ...interface{}) {
	logs.DEBUG().Output(2, logMessage(r, v))
}

// Debugf 相当于调用了 logs.Debugf，外加一些调用者的详细信息
func Debugf(r *http.Request, format string, v ...interface{}) {
	logs.DEBUG().Output(2, logMessagef(r, format, v))
}

// Trace 相当于调用了 logs.Trace，外加一些调用者的详细信息
func Trace(r *http.Request, v ...interface{}) {
	logs.TRACE().Output(2, logMessage(r, v))
}

// Tracef 相当于调用了 logs.Tracef，外加一些调用者的详细信息
func Tracef(r *http.Request, format string, v ...interface{}) {
	logs.TRACE().Output(2, logMessagef(r, format, v))
}

// Warn 相当于调用了 logs.Warn，外加一些调用者的详细信息
func Warn(r *http.Request, v ...interface{}) {
	logs.WARN().Output(2, logMessage(r, v))
}

// Warnf 相当于调用了 logs.Warnf，外加一些调用者的详细信息
func Warnf(r *http.Request, format string, v ...interface{}) {
	logs.WARN().Output(2, logMessagef(r, format, v))
}

// Info 相当于调用了 logs.Info，外加一些调用者的详细信息
func Info(r *http.Request, v ...interface{}) {
	logs.INFO().Output(2, logMessage(r, v))
}

// Infof 相当于调用了 logs.Infof，外加一些调用者的详细信息
func Infof(r *http.Request, format string, v ...interface{}) {
	logs.INFO().Output(2, logMessagef(r, format, v))
}
