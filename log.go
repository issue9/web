// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"

	"github.com/issue9/logs"
)

func message(r *http.Request, v []interface{}) []interface{} {
	// 0 是 message，1 是 Error 等函数，2 才是真正调用 Error 等函数的函数
	v = append(v, "@", r.URL)
	return v
}

// Critical 相当于调用了 logs.Critical，外加一些调用者的详细信息
func Critical(r *http.Request, v ...interface{}) {
	logs.Critical(message(r, v)...)
}

// Criticalf 相当于调用了 logs.Criticalf，外加一些调用者的详细信息
func Criticalf(r *http.Request, format string, v ...interface{}) {
	logs.Criticalf(format, message(r, v)...)
}

// Error 相当于调用了 logs.Error，外加一些调用者的详细信息
func Error(r *http.Request, v ...interface{}) {
	logs.Error(message(r, v)...)
}

// Errorf 相当于调用了 logs.Errorf，外加一些调用者的详细信息
func Errorf(r *http.Request, format string, v ...interface{}) {
	logs.Errorf(format, message(r, v)...)
}

// Debug 相当于调用了 logs.Debug，外加一些调用者的详细信息
func Debug(r *http.Request, v ...interface{}) {
	logs.Debug(message(r, v)...)
}

// Debugf 相当于调用了 logs.Debugf，外加一些调用者的详细信息
func Debugf(r *http.Request, format string, v ...interface{}) {
	logs.Debugf(format, message(r, v)...)
}

// Trace 相当于调用了 logs.Trace，外加一些调用者的详细信息
func Trace(r *http.Request, v ...interface{}) {
	logs.Trace(message(r, v)...)
}

// Tracef 相当于调用了 logs.Tracef，外加一些调用者的详细信息
func Tracef(r *http.Request, format string, v ...interface{}) {
	logs.Tracef(format, message(r, v)...)
}

// Warn 相当于调用了 logs.Warn，外加一些调用者的详细信息
func Warn(r *http.Request, v ...interface{}) {
	logs.Warn(message(r, v)...)
}

// Warnf 相当于调用了 logs.Warnf，外加一些调用者的详细信息
func Warnf(r *http.Request, format string, v ...interface{}) {
	logs.Warnf(format, message(r, v)...)
}

// Info 相当于调用了 logs.Info，外加一些调用者的详细信息
func Info(r *http.Request, v ...interface{}) {
	logs.Info(message(r, v)...)
}

// Infof 相当于调用了 logs.Infof，外加一些调用者的详细信息
func Infof(r *http.Request, format string, v ...interface{}) {
	logs.Infof(format, message(r, v)...)
}
