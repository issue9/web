// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/issue9/logs"
	"github.com/issue9/mux"
	"github.com/issue9/utils"
	"github.com/issue9/web/context"
	"github.com/issue9/web/modules"
)

// 一些错误信息的定义
var (
	ErrAppClosed          = errors.New("当前实例已经关闭")
	ErrConfigDirNotExists = errors.New("配置文件的目录不存在")
)

// BuildHandler 将一个 http.Handler 封装成另一个 http.Handler
type BuildHandler func(http.Handler) http.Handler

// App 保存整个程序的运行环境，方便做整体的调度，比如重启等。
type App struct {
	configDir string
	closed    bool
	router    *mux.Prefix
	builder   BuildHandler

	config  *config
	server  *server
	modules *modules.Modules
}

// New 初始化框架的基本内容。
//
// confDir 指定了配置文件所在的目录，
// 框架默认的两个配置文件都会从此目录下查找。
// confDir 下面必须包含 logs.xml 与 web.json 两个配置文件。
// builder 被用于封装内部的 http.Handler 接口，不需要可以传递空值。
func New(confDir string, builder BuildHandler) (*App, error) {
	if !utils.FileExists(confDir) {
		return nil, ErrConfigDirNotExists
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

	app.config, err = loadConfig(app.File(configFilename))
	if err != nil {
		return err
	}

	app.server, err = newServer(app.config)
	if err != nil {
		return err
	}

	// router
	u, err := url.Parse(app.config.Root)
	if err != nil {
		return err
	}
	app.router = app.server.mux.Prefix(u.Path)

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
		return app.server.run(nil)
	}

	return app.server.run(app.builder(app.Router().Mux()))
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
		return app.server.shutdown(timeout)
	}

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

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则返回 nil
func (app *App) NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	conf := app.config
	ctx, err := context.New(w, r, conf.OutputEncoding, conf.OutputCharset, conf.Strict)

	switch {
	case err == context.ErrUnsupportedContentType:
		context.RenderStatus(w, http.StatusUnsupportedMediaType)
		return nil
	case err == context.ErrClientNotAcceptable:
		context.RenderStatus(w, http.StatusNotAcceptable)
		return nil
	}

	return ctx
}
