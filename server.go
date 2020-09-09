// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/issue9/logs/v2"
	"github.com/issue9/logs/v2/config"
	"github.com/issue9/middleware"
	"github.com/issue9/mux/v2"
	"github.com/issue9/scheduled"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v2"

	context2 "github.com/issue9/web/context"
	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/internal/webconfig"
	"github.com/issue9/web/server"
)

// 两个配置文件的名称。
//
// 在 Classic 中，采用了这两个文件名作为日志和框架的配置文件名。
const (
	LogsFilename   = "logs.xml"
	ConfigFilename = "web.yaml"
)

var (
	defaultConfigs *ConfigManager
	defaultServer  *server.Server
)

// Classic 初始化一个可运行的框架环境
//
// dir 为配置文件的根目录
func Classic(dir string, get context2.BuildResultFunc) error {
	mgr, err := NewConfigManager(dir)
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

	if err = Init(mgr, ConfigFilename, LogsFilename, get); err != nil {
		return err
	}

	lc := &config.Config{}
	if err = mgr.LoadFile(LogsFilename, lc); err != nil {
		return err
	}
	if err = Server().Builder().Logs().Init(lc); err != nil {
		return err
	}

	/*err = AddCompresses(map[string]compress.WriterFunc{
		"gzip":    compress.NewGzip,
		"deflate": compress.NewDeflate,
	})
	if err != nil {
		return err
	}*/

	err = Builder().AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/json":    json.Unmarshal,
		"multipart/form-data": nil,
	})
	if err != nil {
		return err
	}

	return Builder().AddMarshals(map[string]mimetype.MarshalFunc{
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
func Init(mgr *ConfigManager, configFilename, logsFilename string, get context2.BuildResultFunc) error {
	if defaultServer != nil {
		panic("不能重复调用 Init")
	}

	if configFilename == "" {
		panic("参数 configFilename 不能为空")
	}
	if logsFilename == "" {
		panic("参数 logsFilename 不能为空")
	}

	lc := &config.Config{}
	if err := mgr.LoadFile(LogsFilename, lc); err != nil {
		return err
	}
	l := logs.New()
	if err := l.Init(lc); err != nil {
		return err
	}

	webconf := &webconfig.WebConfig{}
	err := mgr.LoadFile(configFilename, webconf)
	if err != nil {
		return err
	}

	defaultConfigs = mgr
	builder := context2.NewServer(l, mux.New(false, false, false, nil, nil).Prefix(""), get)
	defaultServer, err = server.New(webconf, builder)
	return err
}

// Server 返回 defaultServer 实例
func Server() *server.Server {
	return defaultServer
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

		if err := Server().Shutdown(ctx); err != nil {
			Server().Builder().Logs().Error(err)
		}
		Server().Builder().Logs().Flush() // 保证内容会被正常输出到日志。
	}()
}

// AddMiddlewares 设置全局的中间件，可多次调用。
func AddMiddlewares(m middleware.Middleware) {
	Server().AddMiddlewares(m)
}

// IsDebug 是否处在调试模式
func IsDebug() bool {
	return Server().IsDebug()
}

// Mux 返回 mux.Mux 实例。
func Mux() *mux.Mux {
	return Server().Mux()
}

// Builder 返回 context.Server
func Builder() *context2.Server {
	return Server().Builder()
}

// Serve 执行监听程序。
func Serve() error {
	return Server().Run()
}

// Close 关闭服务。
//
// 无论配置文件如果设置，此函数都是直接关闭服务，不会等待。
func Close() error {
	return Server().Close()
}

// Shutdown 关闭所有服务。
//
// 根据配置文件中的配置项，决定当前是直接关闭还是延时之后关闭。
func Shutdown(ctx context.Context) error {
	return Server().Shutdown(ctx)
}

// URL 构建一条完整 URL
func URL(path string) string {
	return Server().URL(path)
}

// Path 构建 URL 的 Path 部分
func Path(path string) string {
	return Server().Path(path)
}

// Services 返回所有的服务列表
func Services() []*server.Service {
	return Server().Services()
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
	Server().Builder().AddMessages(status, messages)
}

// Messages 获取所有的错误消息代码
//
// p 用于返回特定语言的内容。如果为空，则表示返回原始值。
func Messages(p *message.Printer) map[int]string {
	return Server().Builder().Messages(p)
}

// Scheduled 获取 scheduled.Server 实例
func Scheduled() *scheduled.Server {
	return Server().Scheduled()
}

// Schedulers 返回所有的计划任务
func Schedulers() []*scheduled.Job {
	return Scheduled().Jobs()
}

// Location 返回当前配置文件中指定的时区信息
func Location() *time.Location {
	return Server().Location()
}

// Now 返回当前时间
//
// 与 time.Now() 的区别在于 Now() 基于当前时区
func Now() time.Time {
	return time.Now().In(Location())
}

// ParseTime 分析时间格式，基于当前时间
func ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, Location())
}

// Uptime 启动的时间
//
// 时区信息与配置文件中的相同
func Uptime() time.Time {
	return Server().Builder().Uptime()
}
