// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/issue9/middleware"
	"github.com/issue9/mux/v2"
	"golang.org/x/text/message"

	"github.com/issue9/web/app"
	"github.com/issue9/web/context"
	"github.com/issue9/web/mimetype"
	"github.com/issue9/web/module"
)

var defaultApp *app.App

// Init 初始化整个应用环境
//
// 重复调用会直接 panic
func Init(dir string) (err error) {
	if defaultApp != nil {
		panic("不能重复调用 Init")
	}

	defaultApp, err = app.New(dir)
	return
}

// App 返回 defaultApp 实例
func App() *app.App {
	return defaultApp
}

// Grace 指定触发 Shutdown() 的信号，若为空，则任意信号都触发。
//
// 多次调用，则每次指定的信号都会起作用，如果由传递了相同的值，
// 则有可能多次触发 Shutdown()。
//
// NOTE: 传递空值，与不调用，其结果是不同的。
// 若是不调用，则不会处理任何信号；若是传递空值调用，则是处理任何要信号。
func Grace(sig ...os.Signal) {
	app.Grace(defaultApp, sig...)
}

// AddMiddlewares 设置全局的中间件，可多次调用。
func AddMiddlewares(m middleware.Middleware) {
	defaultApp.AddMiddlewares(m)
}

// IsDebug 是否处在调试模式
func IsDebug() bool {
	return defaultApp.IsDebug()
}

// Mux 返回 mux.Mux 实例。
func Mux() *mux.Mux {
	return defaultApp.Mux()
}

// Mimetypes 返回 mimetype.Mimetypes
func Mimetypes() *mimetype.Mimetypes {
	return defaultApp.Mimetypes()
}

// Serve 运行路由，执行监听程序。
func Serve() error {
	return defaultApp.Serve()
}

// InitModules 初始化指定标签的模块
func InitModules(tag string) error {
	return defaultApp.InitModules(tag)
}

// Close 关闭服务。
//
// 无论配置文件如果设置，此函数都是直接关闭服务，不会等待。
func Close() error {
	return defaultApp.Close()
}

// Shutdown 关闭所有服务。
//
// 根据配置文件中的配置项，决定当前是直接关闭还是延时之后关闭。
func Shutdown() error {
	return defaultApp.Shutdown()
}

// URL 构建一条完整 URL
func URL(path string) string {
	return defaultApp.URL(path)
}

// Path 构建 URL 的 Path 部分
func Path(path string) string {
	return defaultApp.Path(path)
}

// Modules 当前系统使用的所有模块信息
func Modules() []*module.Module {
	return defaultApp.Modules()
}

// Tags 获取所有的子模块名称
func Tags() []string {
	return defaultApp.Tags()
}

// RegisterOnShutdown 注册在关闭服务时需要执行的操作。
func RegisterOnShutdown(f func()) {
	defaultApp.RegisterOnShutdown(f)
}

// NewModule 注册一个模块
func NewModule(name, desc string, deps ...string) *Module {
	return defaultApp.NewModule(name, desc, deps...)
}

// File 获取配置目录下的文件。
func File(path string) string {
	return defaultApp.Config().File(path)
}

// LoadFile 加载指定的配置文件内容到 v 中
func LoadFile(path string, v interface{}) error {
	return defaultApp.Config().LoadFile(path, v)
}

// Load 加载指定的配置文件内容到 v 中
func Load(r io.Reader, typ string, v interface{}) error {
	return defaultApp.Config().Load(r, typ, v)
}

// NewMessages 添加新的错误消息代码
func NewMessages(status int, messages map[int]string) {
	defaultApp.NewMessages(status, messages)
}

// Messages 获取所有的错误消息代码
//
// 如果指定 p 的值，则返回本地化的消息内容。
func Messages(p *message.Printer) map[int]string {
	return defaultApp.Messages(p)
}

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则 panic
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return context.New(w, r, defaultApp)
}

// INFO 获取 INFO 级别的 log.Logger 实例，在未指定 info 级别的日志时，该实例返回一个 nil。
func INFO() *log.Logger {
	return defaultApp.Logs().INFO()
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
	return defaultApp.Logs().DEBUG()
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
	return defaultApp.Logs().TRACE()
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
	return defaultApp.Logs().WARN()
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
	return defaultApp.Logs().ERROR()
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
	return defaultApp.Logs().CRITICAL()
}

// Critical 相当于 CRITICAL().Println(v...)的简写方式
func Critical(v ...interface{}) {
	CRITICAL().Output(2, fmt.Sprintln(v...))
}

// Criticalf 相当于 CRITICAL().Printf(format, v...) 的简写方式
func Criticalf(format string, v ...interface{}) {
	CRITICAL().Output(2, fmt.Sprintf(format, v...))
}

// Fatal 输出错误信息，然后退出程序。
func Fatal(code int, v ...interface{}) {
	defaultApp.Logs().Fatal(code, v...)
}

// Fatalf 输出错误信息，然后退出程序。
func Fatalf(code int, format string, v ...interface{}) {
	defaultApp.Logs().Fatalf(code, format, v...)
}

// Panic 输出错误信息，然后触发 panic。
func Panic(v ...interface{}) {
	defaultApp.Logs().Panic(v...)
}

// Panicf 输出错误信息，然后触发 panic。
func Panicf(format string, v ...interface{}) {
	defaultApp.Logs().Panicf(format, v...)
}

// FlushLogs 立即输出所有的日志信息。
func FlushLogs() {
	defaultApp.Logs().Flush()
}
