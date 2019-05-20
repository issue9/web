// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/issue9/config"
	lconf "github.com/issue9/logs/v2/config"
	"github.com/issue9/middleware"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/middleware/recovery/errorhandler"
	"github.com/issue9/mux/v2"
	"github.com/issue9/scheduled"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/app"
	"github.com/issue9/web/context"
	"github.com/issue9/web/internal/webconfig"
	"github.com/issue9/web/mimetype"
	"github.com/issue9/web/module"
)

// 两个配置文件的名称。
//
// 在 Classic 中，采用了这两个文件名作为日志和框架的配置文件名。
const (
	LogsFilename   = "logs.xml"
	ConfigFilename = "web.yaml"
)

var (
	defaultConfigs *config.Manager
	defaultApp     *app.App
)

// Classic 初始化一个可运行的框架环境
//
// dir 为配置文件的根目录
func Classic(dir string, get app.GetResultFunc) error {
	mgr, err := config.NewManager(dir)
	if err != nil {
		return err
	}

	if err = mgr.AddUnmarshal(yaml.Unmarshal, ".yaml", ".yml"); err != nil {
		return err
	}
	if err = mgr.AddUnmarshal(json.Unmarshal, ".json"); err != nil {
		return err
	}
	if err = mgr.AddUnmarshal(xml.Unmarshal, ".xml"); err != nil {
		return err
	}

	if err = Init(mgr, ConfigFilename, get); err != nil {
		return err
	}

	lc := &lconf.Config{}
	if err = mgr.LoadFile(LogsFilename, lc); err != nil {
		return err
	}
	defaultApp.Logs().Init(lc)

	err = AddCompresses(map[string]compress.WriterFunc{
		"gzip":    compress.NewGzip,
		"deflate": compress.NewDeflate,
	})
	if err != nil {
		return err
	}

	err = Mimetypes().AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/json":    json.Unmarshal,
		"multipart/form-data": nil,
	})
	if err != nil {
		return err
	}

	return Mimetypes().AddMarshals(map[string]mimetype.MarshalFunc{
		"application/json":    json.Marshal,
		"multipart/form-data": nil,
	})
}

// Init 初始化整个应用环境
//
// 构建了一个最基本的服务器运行环境，大部分内容采用默认设置。
// 比如日志为不输出任何内容，如有需要，要调用 Logs() 进行输出通道的设置；
// 也不会解析任意的 content-type 内容的数据，需要通过 Mimetype 进行进一步的设置。
//
// mgr 为配置文件管理工具；
// configFilename 为相对于 mgr 目录下的配置文件地址。
//
// 重复调用会直接 panic
func Init(mgr *config.Manager, configFilename string, get app.GetResultFunc) error {
	if defaultApp != nil {
		panic("不能重复调用 Init")
	}

	if configFilename == "" {
		panic("参数 configFilename 不能为空")
	}

	webconf := &webconfig.WebConfig{}
	err := mgr.LoadFile(configFilename, webconf)
	if err != nil {
		return err
	}

	defaultConfigs = mgr
	defaultApp = app.New(webconf, get)

	modules, err = module.NewModules(defaultApp, webconf.Plugins)
	return err
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

// AddCompresses 添加压缩处理函数
func AddCompresses(m map[string]compress.WriterFunc) error {
	return defaultApp.AddCompresses(m)
}

// AddMiddlewares 设置全局的中间件，可多次调用。
func AddMiddlewares(m middleware.Middleware) {
	defaultApp.After(m)
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

// Services 返回所有的服务列表
func Services() []*app.Service {
	return defaultApp.Services()
}

// Server 获取 http.Server 实例
func Server() *http.Server {
	return defaultApp.Server()
}

// File 获取配置目录下的文件。
func File(path string) string {
	return defaultConfigs.File(path)
}

// LoadFile 加载指定的配置文件内容到 v 中
func LoadFile(path string, v interface{}) error {
	return defaultConfigs.LoadFile(path, v)
}

// Load 加载指定的配置文件内容到 v 中
func Load(r io.Reader, typ string, v interface{}) error {
	return defaultConfigs.Load(r, typ, v)
}

// AddMessages 添加新的错误消息代码
func AddMessages(status int, messages map[int]string) {
	defaultApp.AddMessages(status, messages)
}

// ErrorHandlers 错误处理功能
func ErrorHandlers() *errorhandler.ErrorHandler {
	return defaultApp.ErrorHandlers()
}

// Messages 获取所有的错误消息代码
//
// 如果指定 p 的值，则返回本地化的消息内容。
func Messages(p *message.Printer) map[int]string {
	return defaultApp.Messages(p)
}

// Scheduled 获取 scheduled.Server 实例
func Scheduled() *scheduled.Server {
	return defaultApp.Scheduled()
}

// Schedulers 返回所有的计划任务
func Schedulers() []*scheduled.Job {
	return Scheduled().Jobs()
}

// Location 返回当前配置文件中指定的时区信息
func Location() *time.Location {
	return defaultApp.Location()
}

// Now 返回当前时间。
//
// 与 time.Now() 的区别在于 Now() 基于当前时区
func Now() time.Time {
	return time.Now().In(Location())
}

// ParseTime 分析时间格式，基于当前时间
func ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, Location())
}

// NewContext 生成 *Context 对象，若是出错则 panic
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return context.New(w, r, App())
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
