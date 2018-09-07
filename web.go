// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/issue9/logs"
	"github.com/issue9/middleware"
	"github.com/issue9/mux"

	"github.com/issue9/web/config"
	"github.com/issue9/web/internal/app"
	"github.com/issue9/web/internal/app/webconfig"
	"github.com/issue9/web/module"
)

const (
	configFilename = "web.yaml" // 配置文件的文件名。
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
)

var (
	configDir  string
	defaultApp *app.App
)

// Init 初始化整个应用环境
//
// dir 表示配置文件的目录；
func Init(dir string) (err error) {
	configDir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}

	if defaultApp != nil {
		return errors.New("不能重复调用 Init")
	}

	if err = logs.InitFromXMLFile(File(logsFilename)); err != nil {
		return err
	}

	webconf := &webconfig.WebConfig{}
	if err = config.LoadFile(File(configFilename), webconf); err != nil {
		return err
	}

	defaultApp, err = app.New(webconf)
	return
}

// Grace 指定触发 Shutdown() 的信号，若为空，则任意信号都触发。
//
// 多次调用，则每次指定的信号都会起作用，如果由传递了相同的值，
// 则有可能多次触发 Shutdown()。
//
// NOTE: 传递空值，与不调用，其结果是不同的。
// 若是不调用，则不会处理任何信号；若是传递空值调用，则是处理任何要信号。
func Grace(sig ...os.Signal) {
	go func() {
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, sig...)

		<-signalChannel
		signal.Stop(signalChannel)

		if err := Shutdown(); err != nil {
			logs.Error(err)
			logs.Flush() // 保证内容会被正常输出到日志。
		}
	}()
}

// SetMiddleware 设置一个全局的中间件，多次设置，只有最后一次会启作用。
func SetMiddleware(m middleware.Middleware) {
	defaultApp.SetMiddleware(m)
}

// IsDebug 是否处在调试模式
func IsDebug() bool {
	return defaultApp.Debug()
}

// Handler 将当前实例当作一个 http.Handler 返回。一般用于测试。
// 比如在 httptest.NewServer 中使用。
func Handler() (http.Handler, error) {
	return defaultApp.Handler()
}

// Mux 返回 mux.Mux 实例。
func Mux() *mux.Mux {
	return defaultApp.Mux()
}

// Run 运行路由，执行监听程序。
//
// Deprecated: 由 Serve 代替
func Run() error {
	return Serve()
}

// Serve 运行路由，执行监听程序。
func Serve() error {
	return defaultApp.Serve()
}

// Install 执行指定版本的安装功能
func Install(version string) error {
	return defaultApp.InitModules(version)
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

// File 获取配置目录下的文件。
func File(path ...string) string {
	paths := make([]string, 0, len(path)+1)
	paths = append(paths, configDir)
	return filepath.Join(append(paths, path...)...)
}

// URL 构建一条完整 URL
func URL(path string) string {
	return defaultApp.URL(path)
}

// Modules 当前系统使用的所有模块信息
func Modules() []*module.Module {
	return defaultApp.Modules()
}

// RegisterOnShutdown 注册在关闭服务时需要执行的操作。
func RegisterOnShutdown(f func()) {
	defaultApp.RegisterOnShutdown(f)
}

// LoadConfig 从配置目录中加载数据到对象 v 中。
func LoadConfig(path string, v interface{}) error {
	return config.LoadFile(path, v)
}

// NewModule 注册一个模块
func NewModule(name, desc string, deps ...string) *Module {
	return defaultApp.NewModule(name, desc, deps...)
}
