// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"net/http"
	"os"
	"os/signal"

	"github.com/issue9/logs"
	"github.com/issue9/web/context"
	"github.com/issue9/web/internal/app"
	"github.com/issue9/web/module"
	"github.com/issue9/web/result"
)

var defaultApp *app.App

func grace(s ...os.Signal) {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, s...)

	<-signalChannel
	if err := Shutdown(); err != nil {
		logs.Error(err)
	}
}

// Init 初始化整个应用环境
//
// configDir 表示配置文件的目录；
// m 表示应用于所有路由项的中间件；
// s 表示触发 shutdown 的信号。
// 传递给框架的信号，会触发调用 Shutdown() 操作。
func Init(configDir string, m module.Middleware, s ...os.Signal) error {
	if defaultApp != nil {
		return errors.New("不能重复调用 Init")
	}

	app, err := app.New(configDir, m)
	if err != nil {
		return err
	}

	defaultApp = app

	if len(s) > 0 {
		go grace(s...)
	}

	return nil
}

// IsDebug 是否处在调试模式
func IsDebug() bool {
	return defaultApp.Debug()
}

// Run 运行路由，执行监听程序。
func Run() error {
	return defaultApp.Run()
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

// NewModule 注册一个模块
func NewModule(name, desc string, deps ...string) *module.Module {
	return defaultApp.NewModule(name, desc, deps...)
}

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则返回 nil
func NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	return defaultApp.NewContext(w, r)
}

// NewResult 生成一个 *result.Result 对象
func NewResult(code int) *result.Result {
	return result.New(code, nil)
}

// NewResultWithDetail 声明一个带详细内容的 result.Result 对象
func NewResultWithDetail(code int, fields map[string]string) *result.Result {
	return result.New(code, fields)
}
