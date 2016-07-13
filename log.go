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

func Critical(v ...interface{}) {
	logs.Critical(message(v)...)
}

func Criticalf(format string, v ...interface{}) {
	logs.Criticalf(format, message(v)...)
}

func Error(v ...interface{}) {
	logs.Error(message(v)...)
}

func Errorf(format string, v ...interface{}) {
	logs.Errorf(format, message(v)...)
}

func Debug(v ...interface{}) {
	logs.Debug(message(v)...)
}

func Debugf(format string, v ...interface{}) {
	logs.Debugf(format, message(v)...)
}

func Trace(v ...interface{}) {
	logs.Trace(message(v)...)
}

func Tracef(format string, v ...interface{}) {
	logs.Tracef(format, message(v)...)
}

func Warn(v ...interface{}) {
	logs.Warn(message(v)...)
}

func Warnf(format string, v ...interface{}) {
	logs.Warnf(format, message(v)...)
}

func Info(v ...interface{}) {
	logs.Info(message(v)...)
}

func Infof(format string, v ...interface{}) {
	logs.Infof(format, message(v)...)
}
