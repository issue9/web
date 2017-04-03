// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"time"

	"github.com/issue9/mux"
	"github.com/issue9/web/modules"
)

// Version 当前框架的版本
const Version = "0.9.3+20170403"

var defaultApp *App

// Init 初始化框架的基本内容。
//
// confDir 指定了配置文件所在的目录，框架默认的
// 两个配置文件都会从此目录下查找。
func Init(confDir string) error {
	app, err := NewApp(confDir)
	if err != nil {
		return err
	}

	defaultApp = app
	return nil
}

// Run 运行路由，执行监听程序。
//
// h 表示需要执行的路由处理函数，传递 nil 时，会自动以 Mux() 代替。
// 可以通过以下方式，将一些 http.Handler 实例附加到 Mux() 之上：
//  web.Run(handlers.Host(web.Mux(), "www.caixw.io")
func Run(h http.Handler) error {
	return defaultApp.Run(h)
}

// Restart 重启整个服务。
//
// timeout 表示已有服务的等待时间。
// 若超过该时间，服务还未自动停止的，则会强制停止，若小于或等于 0 则立即重启。
func Restart(timeout time.Duration) error {
	return defaultApp.Restart(timeout)
}

// Shutdown 关闭所有服务。
// 关闭之后不能再调用 Run() 重新运行。
// 若只是想重启服务，只能调用 Restart() 函数。
//
// timeout 表示已有服务的等待时间。
// 若超过该时间，服务还未自动停止的，则会强制停止，若小于或等于 0 则立即重启。
func Shutdown(timeout time.Duration) error {
	return defaultApp.Shutdown(timeout)
}

// File 获取配置目录下的文件。
func File(path string) string {
	return defaultApp.File(path)
}

// Mux 获取 mux.ServeMux 实例。
//
// 通过 Mux 可以添加各类路由项，诸如：
//  Mux().Get("/test", h).
//      Post("/test", h)
func Mux() *mux.ServeMux {
	return defaultApp.Mux()
}

// NewModule 注册一个新的模块。
//
// name 为模块名称；
// init 当前模块的初始化函数；
// deps 模块的依赖模块，这些模块在初始化时，会先于 name 初始化始。
func NewModule(name string, init modules.Init, deps ...string) {
	defaultApp.NewModule(name, init, deps...)
}
