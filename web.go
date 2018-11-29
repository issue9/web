// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"net/http"
	"os"
	"os/signal"

	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware"
	"github.com/issue9/mux"

	"github.com/issue9/web/internal/app"
	"github.com/issue9/web/module"
)

var defaultApp *app.App

// Init 初始化整个应用环境
//
// dir 表示配置文件的目录；
func Init(dir string) (err error) {
	if defaultApp != nil {
		return errors.New("不能重复调用 Init")
	}

	defaultApp, err = app.New(dir)
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

// SetMiddlewares 设置一个全局的中间件，多次设置，只有最后一次会启作用。
func SetMiddlewares(m ...middleware.Middleware) {
	defaultApp.SetMiddlewares(m...)
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
func File(path string) string {
	return defaultApp.File(path)
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
	return defaultApp.LoadFile(path, v)
}

// NewModule 注册一个模块
func NewModule(name, desc string, deps ...string) *Module {
	return defaultApp.NewModule(name, desc, deps...)
}
