// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"net/http"
	"path/filepath"
	"time"

	"github.com/issue9/logs"
	"github.com/issue9/mux"
	"github.com/issue9/utils"
	"github.com/issue9/web/content"
	"github.com/issue9/web/internal/config"
	"github.com/issue9/web/internal/server"
	"github.com/issue9/web/modules"
)

const (
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
	configFilename = "web.json" // 配置文件的文件名。
)

var defaultApp *App

// App 保存整个程序的运行环境，方便做整体的调度，比如重启等。
type App struct {
	configDir string
	config    *config.Config
	server    *server.Server
	content   content.Content
	modules   *modules.Modules
	handler   http.Handler
}

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

// NewApp 初始化框架的基本内容。
//
// confDir 指定了配置文件所在的目录，框架默认的
// 两个配置文件都会从此目录下查找。
func NewApp(confDir string) (*App, error) {
	if !utils.FileExists(confDir) {
		return nil, errors.New("配置文件目录不存在")
	}

	app := &App{
		configDir: confDir,
		modules:   modules.New(),
	}

	if err := app.init(); err != nil {
		return nil, err
	}

	return app, nil
}

// 对其它可重复加载的数据进行初始化。
// 方便在 NewApp 和 Restart 中进行调用。
func (app *App) init() error {
	err := logs.InitFromXMLFile(app.File(logsFilename))
	if err != nil {
		return err
	}

	app.config, err = config.Load(app.File(configFilename))
	if err != nil {
		return err
	}

	app.server, err = server.New(app.config.Server)
	if err != nil {
		return err
	}

	app.content, err = content.New(app.config.Content)
	if err != nil {
		return err
	}

	return nil
}

// File 获取配置目录下的文件。
func (app *App) File(path string) string {
	return filepath.Join(app.configDir, path)
}

// Mux 获取 mux.ServeMux 实例。
//
// 通过 Mux 可以添加各类路由项，诸如：
//  Mux().Get("/test", h).
//      Post("/test", h)
func (app *App) Mux() *mux.ServeMux {
	return app.server.Mux()
}

// NewModule 注册一个新的模块。
//
// name 为模块名称；
// init 当前模块的初始化函数；
// deps 模块的依赖模块，这些模块在初始化时，会先于 name 初始化始。
func (app *App) NewModule(name string, init modules.Init, deps ...string) {
	err := app.modules.New(name, init, deps...)

	// 注册模块时出错，直接退出。
	if err != nil {
		logs.Fatal(err)
	}
}

// Run 运行路由，执行监听程序。
//
// h 表示需要执行的路由处理函数，传递 nil 时，会自动以 App.Mux() 代替。
// 可以通过以下方式，将一些 http.Handler 实例附加到 App.Mux() 之上：
//  app.Run(handlers.Host(app.Mux(), "www.caixw.io")
func (app *App) Run(h http.Handler) error {
	if err := app.modules.Init(); err != nil {
		return err
	}

	app.handler = h
	return app.server.Run(app.handler)
}

// Shutdown 关闭所有服务。
// 若只是想重启服务，只能调用 Restart() 函数。
//
// timeout 表示已有服务的等待时间。
// 若超过该时间，服务还未自动停止的，则会强制停止，若小于或等于 0 则立即重启。
func (app *App) Shutdown(timeout time.Duration) error {
	logs.Flush()

	if app.server != nil {
		return app.server.Shutdown(timeout)
	}

	return nil
}

// Restart 重启整个服务。
// Restart 并不简单地等于 Shutdown() + Run()。
// 重启时，会得新加载配置文件内容；清除路由项；重新调用模块初始化函数。要想保持
// 路由继续启作用，请将路由的初发化工作放到模块的初始化函数中。
//
// timeout 表示已有服务的等待时间。
// 若超过该时间，服务还未自动停止的，则会强制停止，若小于或等于 0 则立即重启。
func (app *App) Restart(timeout time.Duration) error {
	if err := app.Shutdown(timeout); err != nil {
		// NOTE: Shutdown() 始返回的错误信息，并不表示程序出错，无需退出，仅作简单的记录。
		logs.Error(err)
	}

	if err := app.init(); err != nil {
		return err
	}

	if err := app.modules.Reset(); err != nil {
		return err
	}

	go func() {
		logs.Error(app.Run(app.handler))
	}()

	return nil
}
