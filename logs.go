// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// logs.go中的函数是对logs包的重新封装。

package web

import (
	"log"

	"github.com/issue9/logs"
)

// 输出所有的缓存内容。
func Flush() {
	logs.Flush()
}

// 相当于logs.INFO()
func INFO() *log.Logger {
	return logs.INFO()
}

// 相当于logs.Info()
func Info(v ...interface{}) {
	logs.Info(v...)
}

// 相当于logs.Infof()
func Infof(format string, v ...interface{}) {
	logs.Infof(format, v...)
}

// 相当于logs.DEBUG()
func DEBUG() *log.Logger {
	return logs.DEBUG()
}

// 相当于logs.Debug()
func Debug(v ...interface{}) {
	logs.Debug(v...)
}

// 相当于logs.Debugf()
func Debugf(format string, v ...interface{}) {
	logs.Debugf(format, v...)
}

// 相当于logs.TRACE()
func TRACE() *log.Logger {
	return logs.TRACE()
}

// 相当于logs.Trace()
func Trace(v ...interface{}) {
	logs.Trace(v...)
}

// 相当于logs.Tracef()
func Tracef(format string, v ...interface{}) {
	logs.Tracef(format, v...)
}

// 相当于logs.WARN()
func WARN() *log.Logger {
	return logs.WARN()
}

// 相当于logs.Warn()
func Warn(v ...interface{}) {
	logs.Warn(v...)
}

// 相当于logs.Warnf()
func Warnf(format string, v ...interface{}) {
	logs.Warnf(format, v...)
}

// 相当于logs.ERROR()
func ERROR() *log.Logger {
	return logs.ERROR()
}

// 相当于logs.Error()
func Error(v ...interface{}) {
	logs.Error(v...)
}

// 相当于logs.Errorf()
func Errorf(format string, v ...interface{}) {
	logs.Errorf(format, v...)
}

// 相当于logs.CRITICAL()
func CRITICAL() *log.Logger {
	return logs.CRITICAL()
}

// 相当于logs.Critical()
func Critical(v ...interface{}) {
	logs.Critical(v...)
}

// 相当于logs.Criticalf()
func Criticalf(format string, v ...interface{}) {
	logs.Criticalf(format, v...)
}

// 相当于logs.All()
func All(v ...interface{}) {
	logs.All(v...)
}

// 相当于logs.Allf()
func Allf(format string, v ...interface{}) {
	logs.Allf(format, v...)
}

// 输出错误信息，然后退出程序。
func Fatal(v ...interface{}) {
	logs.Fatal(v...)
}

// 输出错误信息，然后退出程序。
func Fatalf(format string, v ...interface{}) {
	logs.Fatalf(format, v...)
}

// 输出错误信息，然后触发panic。
func Panic(v ...interface{}) {
	logs.Panic(v...)
}

// 输出错误信息，然后触发panic。
func Panicf(format string, v ...interface{}) {
	logs.Panicf(format, v...)
}
