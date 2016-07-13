// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"runtime"

	"github.com/issue9/logs"
)

func message(v []interface{}) []interface{} {
	// 0 是 message，1 是 Error 等函数，2 才是真正调用 Error 等函数的函数
	if pc, _, _, ok := runtime.Caller(2); ok {
		v = append(v, "@", runtime.FuncForPC(pc).Name())
	}
	return v
}

// Critical 相当于调用了 logs.Critical，外加一些调用者的详细信息
func Critical(v ...interface{}) {
	logs.Critical(message(v)...)
}

// Criticalf 相当于调用了 logs.Criticalf，外加一些调用者的详细信息
func Criticalf(format string, v ...interface{}) {
	logs.Criticalf(format, message(v)...)
}

// Error 相当于调用了 logs.Error，外加一些调用者的详细信息
func Error(v ...interface{}) {
	logs.Error(message(v)...)
}

// Errorf 相当于调用了 logs.Errorf，外加一些调用者的详细信息
func Errorf(format string, v ...interface{}) {
	logs.Errorf(format, message(v)...)
}

// Debug 相当于调用了 logs.Debug，外加一些调用者的详细信息
func Debug(v ...interface{}) {
	logs.Debug(message(v)...)
}

// Debugf 相当于调用了 logs.Debugf，外加一些调用者的详细信息
func Debugf(format string, v ...interface{}) {
	logs.Debugf(format, message(v)...)
}

// Trace 相当于调用了 logs.Trace，外加一些调用者的详细信息
func Trace(v ...interface{}) {
	logs.Trace(message(v)...)
}

// Tracef 相当于调用了 logs.Tracef，外加一些调用者的详细信息
func Tracef(format string, v ...interface{}) {
	logs.Tracef(format, message(v)...)
}

// Warn 相当于调用了 logs.Warn，外加一些调用者的详细信息
func Warn(v ...interface{}) {
	logs.Warn(message(v)...)
}

// Warnf 相当于调用了 logs.Warnf，外加一些调用者的详细信息
func Warnf(format string, v ...interface{}) {
	logs.Warnf(format, message(v)...)
}

// Info 相当于调用了 logs.Info，外加一些调用者的详细信息
func Info(v ...interface{}) {
	logs.Info(message(v)...)
}

// Infof 相当于调用了 logs.Infof，外加一些调用者的详细信息
func Infof(format string, v ...interface{}) {
	logs.Infof(format, message(v)...)
}
