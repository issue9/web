// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"net/http"
	"net/url"
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

// ErrAppClosed 表示当前的 app 实例已经关闭。
var ErrAppClosed = errors.New("当前实例已经关闭")

var defaultApp *App

// BuildHandler 将一个 http.Handler 封装成另一个 http.Handler
type BuildHandler func(http.Handler) http.Handler

// App 保存整个程序的运行环境，方便做整体的调度，比如重启等。
type App struct {
	configDir string
	closed    bool
	router    *mux.Prefix
	builder   BuildHandler

	config  *config.Config
	server  *server.Server
	content content.Content
	modules *modules.Modules
}

// Init 初始化框架的基本内容。参数说明可参考 NewApp() 的文档。
func Init(confDir string, builder BuildHandler) error {
	app, err := NewApp(confDir, builder)
	if err != nil {
		return err
	}

	defaultApp = app
	return nil
}

// Run 运行路由，执行监听程序，具体说明可参考 App.Run()。
func Run() error {
	return defaultApp.Run()
}

// Restart 重启整个服务，具体说明可参考 App.Restart()。
func Restart(timeout time.Duration) error {
	return defaultApp.Restart(timeout)
}

// Shutdown 关闭所有服务，具体说明可参考 App.Shutdown()
func Shutdown(timeout time.Duration) error {
	return defaultApp.Shutdown(timeout)
}

// File 获取配置目录下的文件，具体说明可参考 App.File()。
func File(path string) string {
	return defaultApp.File(path)
}

// Router 获取操作路由的接口，为一个 mux.Prefix 实例，具体接口说明可参考 issue9/mux 包。
func Router() *mux.Prefix {
	return defaultApp.Router()
}

// URL 构建一条基于 Config.Root 的完整 URL
func URL(path string) string {
	return defaultApp.URL(path)
}

// Module 注册一个新的模块，具体说明可参考 App.Mux()。
func Module(name string, init modules.InitFunc, deps ...string) {
	defaultApp.Module(name, init, deps...)
}

// NewApp 初始化框架的基本内容。
//
// confDir 指定了配置文件所在的目录，
// 框架默认的两个配置文件都会从此目录下查找。
// confDir 下面必须包含 logs.xml 与 web.json 两个配置文件。
// builder 被用于封装内部的 http.Handler 接口，不需要可以传递空值。
func NewApp(confDir string, builder BuildHandler) (*App, error) {
	if !utils.FileExists(confDir) {
		return nil, errors.New("配置文件目录不存在")
	}

	app := &App{
		configDir: confDir,
		modules:   modules.New(),
		builder:   builder,
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

	// router
	u, err := url.Parse(app.config.Root)
	if err != nil {
		return err
	}
	app.router = app.server.Mux().Prefix(u.Path)

	return nil
}

// Run 运行路由，执行监听程序。
func (app *App) Run() error {
	if app.closed {
		return ErrAppClosed
	}

	if err := app.modules.Init(); err != nil {
		return err
	}

	if app.builder == nil {
		return app.server.Run(nil)
	}

	return app.server.Run(app.builder(app.Router().Mux()))
}

// Shutdown 关闭所有服务，之后 app 实例将不可再用，
// 若只是想重启服务，只能调用 Restart() 函数。
//
// timeout 表示已有服务的等待时间。
// 若超过该时间，服务还未自动停止的，则会强制停止，若小于或等于 0 则立即重启。
func (app *App) Shutdown(timeout time.Duration) error {
	logs.Flush()
	app.closed = true

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
	if app.closed {
		return ErrAppClosed
	}

	if app.server != nil {
		// NOTE: Shutdown() 始终返回的错误信息，并不表示程序出错，无需退出，仅作简单的记录。
		logs.Error(app.server.Shutdown(timeout))
	}

	if err := app.init(); err != nil {
		return err
	}

	if err := app.modules.Reset(); err != nil {
		return err
	}

	go func() {
		logs.Error(app.Run())
	}()

	return nil
}

// File 获取相对于配置目录下的文件。
func (app *App) File(path string) string {
	return filepath.Join(app.configDir, path)
}

// Router 获取操作路由的接口，为一个 mux.Prefix 实例，具体接口说明可参考 issue9/mux 包。
//
// 通过 Router 可以添加各类路由项，诸如：
//  Router().Get("/test", h).
//      Post("/test", h)
func (app *App) Router() *mux.Prefix {
	return app.router
}

// Module 注册一个新的模块。
//
// name 为模块名称；
// init 当前模块的初始化函数；
// deps 模块的依赖模块，这些模块在初始化时，会先于 name 初始化始。
func (app *App) Module(name string, init modules.InitFunc, deps ...string) {
	err := app.modules.New(name, init, deps...)

	// 注册模块时出错，直接退出。
	if err != nil {
		logs.Fatal(err)
	}
}

// URL 构建一条基于 Config.Root 的完整 URL
func (app *App) URL(path string) string {
	if len(path) == 0 {
		return app.config.Root
	}

	if path[0] != '/' {
		path = "/" + path
	}
	return app.config.Root + path
}
