// SPDX-License-Identifier: MIT

package web

import (
	"fmt"
	"log"
	"os"
)

// INFO 获取 INFO 级别的 log.Logger 实例，在未指定 info 级别的日志时，该实例返回一个 nil。
func INFO() *log.Logger {
	return Server().Logs().INFO()
}

// Info 相当于 INFO().Println(v...) 的简写方式
// Info 函数默认是带换行符的，若需要不带换行符的，请使用 DEBUG().Print() 函数代替。
// 其它相似函数也有类型功能。
func Info(v ...interface{}) {
	INFO().Output(2, fmt.Sprintln(v...))
}

// Infof 相当于 INFO().Printf(format, v...) 的简写方式
func Infof(format string, v ...interface{}) {
	INFO().Output(2, fmt.Sprintf(format, v...))
}

// DEBUG 获取 DEBUG 级别的 log.Logger 实例，在未指定 debug 级别的日志时，该实例返回一个 nil。
func DEBUG() *log.Logger {
	return Server().Logs().DEBUG()
}

// Debug 相当于 DEBUG().Println(v...) 的简写方式
func Debug(v ...interface{}) {
	DEBUG().Output(2, fmt.Sprintln(v...))
}

// Debugf 相当于 DEBUG().Printf(format, v...) 的简写方式
func Debugf(format string, v ...interface{}) {
	DEBUG().Output(2, fmt.Sprintf(format, v...))
}

// TRACE 获取 TRACE 级别的 log.Logger 实例，在未指定 trace 级别的日志时，该实例返回一个 nil。
func TRACE() *log.Logger {
	return Server().Logs().TRACE()
}

// Trace 相当于 TRACE().Println(v...) 的简写方式
func Trace(v ...interface{}) {
	TRACE().Output(2, fmt.Sprintln(v...))
}

// Tracef 相当于 TRACE().Printf(format, v...) 的简写方式
func Tracef(format string, v ...interface{}) {
	TRACE().Output(2, fmt.Sprintf(format, v...))
}

// WARN 获取 WARN 级别的 log.Logger 实例，在未指定 warn 级别的日志时，该实例返回一个 nil。
func WARN() *log.Logger {
	return Server().Logs().WARN()
}

// Warn 相当于 WARN().Println(v...) 的简写方式
func Warn(v ...interface{}) {
	WARN().Output(2, fmt.Sprintln(v...))
}

// Warnf 相当于 WARN().Printf(format, v...) 的简写方式
func Warnf(format string, v ...interface{}) {
	WARN().Output(2, fmt.Sprintf(format, v...))
}

// ERROR 获取 ERROR 级别的 log.Logger 实例，在未指定 error 级别的日志时，该实例返回一个 nil。
func ERROR() *log.Logger {
	return Server().Logs().ERROR()
}

// Error 相当于 ERROR().Println(v...) 的简写方式
func Error(v ...interface{}) {
	ERROR().Output(2, fmt.Sprintln(v...))
}

// Errorf 相当于 ERROR().Printf(format, v...) 的简写方式
func Errorf(format string, v ...interface{}) {
	ERROR().Output(2, fmt.Sprintf(format, v...))
}

// CRITICAL 获取 CRITICAL 级别的 log.Logger 实例，在未指定 critical 级别的日志时，该实例返回一个 nil。
func CRITICAL() *log.Logger {
	return Server().Logs().CRITICAL()
}

// Critical 相当于 CRITICAL().Println(v...)的简写方式
func Critical(v ...interface{}) {
	CRITICAL().Output(2, fmt.Sprintln(v...))
}

// Criticalf 相当于 CRITICAL().Printf(format, v...) 的简写方式
func Criticalf(format string, v ...interface{}) {
	CRITICAL().Output(2, fmt.Sprintf(format, v...))
}

// 输出 msg 到所有日志通道，并立即刷新内容。
func all(msg string) {
	INFO().Output(3, msg)
	DEBUG().Output(3, msg)
	WARN().Output(3, msg)
	TRACE().Output(3, msg)
	ERROR().Output(3, msg)
	CRITICAL().Output(3, msg)
	Server().Logs().Flush()
}

// Fatal 输出错误信息，然后退出程序。
func Fatal(code int, v ...interface{}) {
	all(fmt.Sprint(v...))
	os.Exit(code)
}

// Fatalf 输出错误信息，然后退出程序。
func Fatalf(code int, format string, v ...interface{}) {
	all(fmt.Sprintf(format, v...))
	os.Exit(code)
}

// Panic 输出错误信息，然后触发 panic。
func Panic(v ...interface{}) {
	msg := fmt.Sprint(v...)
	all(msg)
	panic(msg)
}

// Panicf 输出错误信息，然后触发 panic。
func Panicf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	all(msg)
	panic(msg)
}
