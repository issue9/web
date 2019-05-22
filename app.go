// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"os/signal"
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
	if err = App().Logs().Init(lc); err != nil {
		return err
	}

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
func Grace(dur time.Duration, sig ...os.Signal) {
	go func() {
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, sig...)

		<-signalChannel
		signal.Stop(signalChannel)
		close(signalChannel)

		ctx, c := context.WithTimeout(context.Background(), dur)
		defer c()

		if err := App().Shutdown(ctx); err != nil {
			App().Logs().Error(err)
		}
		App().Logs().Flush() // 保证内容会被正常输出到日志。
	}()
}

// AddCompresses 添加压缩处理函数
func AddCompresses(m map[string]compress.WriterFunc) error {
	return App().AddCompresses(m)
}

// AddMiddlewares 设置全局的中间件，可多次调用。
func AddMiddlewares(m middleware.Middleware) {
	App().AddMiddlewares(m)
}

// IsDebug 是否处在调试模式
func IsDebug() bool {
	return App().IsDebug()
}

// Mux 返回 mux.Mux 实例。
func Mux() *mux.Mux {
	return App().Mux()
}

// Mimetypes 返回 mimetype.Mimetypes
func Mimetypes() *mimetype.Mimetypes {
	return App().Mimetypes()
}

// Serve 执行监听程序。
func Serve() error {
	return App().Run()
}

// Close 关闭服务。
//
// 无论配置文件如果设置，此函数都是直接关闭服务，不会等待。
func Close() error {
	return App().Close()
}

// Shutdown 关闭所有服务。
//
// 根据配置文件中的配置项，决定当前是直接关闭还是延时之后关闭。
func Shutdown(ctx context.Context) error {
	return App().Shutdown(ctx)
}

// URL 构建一条完整 URL
func URL(path string) string {
	return App().URL(path)
}

// Path 构建 URL 的 Path 部分
func Path(path string) string {
	return App().Path(path)
}

// Services 返回所有的服务列表
func Services() []*app.Service {
	return App().Services()
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
	App().AddMessages(status, messages)
}

// ErrorHandlers 错误处理功能
func ErrorHandlers() *errorhandler.ErrorHandler {
	return App().ErrorHandlers()
}

// Messages 获取所有的错误消息代码
//
// p 用于返回特定语言的内容。如果为空，则表示返回原始值。
func Messages(p *message.Printer) map[int]string {
	return App().Messages(p)
}

// Scheduled 获取 scheduled.Server 实例
func Scheduled() *scheduled.Server {
	return App().Scheduled()
}

// Schedulers 返回所有的计划任务
func Schedulers() []*scheduled.Job {
	return Scheduled().Jobs()
}

// Location 返回当前配置文件中指定的时区信息
func Location() *time.Location {
	return App().Location()
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
