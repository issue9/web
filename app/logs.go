// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"fmt"
	"log"

	"github.com/issue9/logs/v2"
)

// FlushLogs 输出所有的缓存内容。
func (app *App) FlushLogs() {
	app.logs.Flush()
}

// INFO 获取 INFO 级别的 log.Logger 实例，在未指定 info 级别的日志时，该实例返回一个 nil。
func (app *App) INFO() *log.Logger {
	return app.logs.INFO()
}

// Info 相当于 INFO().Println(v...) 的简写方式
// Info 函数默认是带换行符的，若需要不带换行符的，请使用 DEBUG().Print() 函数代替。
// 其它相似函数也有类型功能。
func (app *App) Info(v ...interface{}) {
	app.logs.INFO().Output(2, fmt.Sprintln(v...))
}

// Infof 相当于 INFO().Printf(format, v...) 的简写方式
func (app *App) Infof(format string, v ...interface{}) {
	app.logs.INFO().Output(2, fmt.Sprintf(format, v...))
}

// DEBUG 获取 DEBUG 级别的 log.Logger 实例，在未指定 debug 级别的日志时，该实例返回一个 nil。
func (app *App) DEBUG() *log.Logger {
	return app.logs.DEBUG()
}

// Debug 相当于 DEBUG().Println(v...) 的简写方式
func (app *App) Debug(v ...interface{}) {
	app.logs.DEBUG().Output(2, fmt.Sprintln(v...))
}

// Debugf 相当于 DEBUG().Printf(format, v...) 的简写方式
func (app *App) Debugf(format string, v ...interface{}) {
	app.logs.DEBUG().Output(2, fmt.Sprintf(format, v...))
}

// TRACE 获取 TRACE 级别的 log.Logger 实例，在未指定 trace 级别的日志时，该实例返回一个 nil。
func (app *App) TRACE() *log.Logger {
	return app.logs.TRACE()
}

// Trace 相当于 TRACE().Println(v...) 的简写方式
func (app *App) Trace(v ...interface{}) {
	app.logs.TRACE().Output(2, fmt.Sprintln(v...))
}

// Tracef 相当于 TRACE().Printf(format, v...) 的简写方式
func (app *App) Tracef(format string, v ...interface{}) {
	app.logs.TRACE().Output(2, fmt.Sprintf(format, v...))
}

// WARN 获取 WARN 级别的 log.Logger 实例，在未指定 warn 级别的日志时，该实例返回一个 nil。
func (app *App) WARN() *log.Logger {
	return app.logs.WARN()
}

// Warn 相当于 WARN().Println(v...) 的简写方式
func (app *App) Warn(v ...interface{}) {
	app.logs.WARN().Output(2, fmt.Sprintln(v...))
}

// Warnf 相当于 WARN().Printf(format, v...) 的简写方式
func (app *App) Warnf(format string, v ...interface{}) {
	app.logs.WARN().Output(2, fmt.Sprintf(format, v...))
}

// ERROR 获取 ERROR 级别的 log.Logger 实例，在未指定 error 级别的日志时，该实例返回一个 nil。
func (app *App) ERROR() *log.Logger {
	return app.logs.ERROR()
}

// Error 相当于 ERROR().Println(v...) 的简写方式
func (app *App) Error(v ...interface{}) {
	app.logs.ERROR().Output(2, fmt.Sprintln(v...))
}

// Errorf 相当于 ERROR().Printf(format, v...) 的简写方式
func (app *App) Errorf(format string, v ...interface{}) {
	app.logs.ERROR().Output(2, fmt.Sprintf(format, v...))
}

// CRITICAL 获取 CRITICAL 级别的 log.Logger 实例，在未指定 critical 级别的日志时，该实例返回一个 nil。
func (app *App) CRITICAL() *log.Logger {
	return app.logs.CRITICAL()
}

// Critical 相当于 CRITICAL().Println(v...)的简写方式
func (app *App) Critical(v ...interface{}) {
	app.logs.CRITICAL().Output(2, fmt.Sprintln(v...))
}

// Criticalf 相当于 CRITICAL().Printf(format, v...) 的简写方式
func (app *App) Criticalf(format string, v ...interface{}) {
	app.logs.CRITICAL().Output(2, fmt.Sprintf(format, v...))
}

// Logs 获取 logs.Logs 实例
func (app *App) Logs() *logs.Logs {
	return app.logs
}

// Fatal 输出错误信息，然后退出程序。
func (app *App) Fatal(code int, v ...interface{}) {
	app.logs.Fatal(code, v...)
}

// Fatalf 输出错误信息，然后退出程序。
func (app *App) Fatalf(code int, format string, v ...interface{}) {
	app.logs.Fatalf(code, format, v...)
}

// Panic 输出错误信息，然后触发 panic。
func (app *App) Panic(v ...interface{}) {
	app.logs.Panic(v...)
}

// Panicf 输出错误信息，然后触发 panic。
func (app *App) Panicf(format string, v ...interface{}) {
	app.logs.Panicf(format, v...)
}
