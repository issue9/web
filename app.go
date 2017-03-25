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

// App 保存整个程序的运行环境，方便做整体的调度，比如重启等。
type App struct {
	configDir string
	config    *config.Config
	server    *server.Server
	content   content.Content
	modules   *modules.Modules
	handler   http.Handler
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

	err := logs.InitFromXMLFile(app.File(logsFilename))
	if err != nil {
		return nil, err
	}

	app.config, err = config.Load(app.File(configFilename))
	if err != nil {
		return nil, err
	}

	app.server, err = server.New(app.config.Server)
	if err != nil {
		return nil, err
	}

	app.content, err = content.New(app.config.Content)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// File 获取配置目录下的文件。
func (app *App) File(path string) string {
	return filepath.Join(app.configDir, path)
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
func (app *App) Run(h http.Handler) error {
	if err := app.modules.Init(); err != nil {
		return err
	}

	app.handler = h
	return app.server.Run(app.handler)
}

// Shutdown 关闭所有服务。
// 关闭之后不能再调用 Run() 重新运行。
// 若只是想重启服务，只能调用 Restart() 函数。
//
// timeout 表示已有服务的等待时间。
// 若超过该时间，服务还未自动停止的，则会强制停止，若小于或等于 0 则立即重启。
func (app *App) Shutdown(timeout time.Duration) error {
	logs.Flush()

	if app.server != nil {
		if err := app.server.Shutdown(timeout); err != nil {
			return err
		}
	}

	return nil
}

// Restart 重启整个服务。
//
// timeout 表示已有服务的等待时间。
// 若超过该时间，服务还未自动停止的，则会强制停止，若小于或等于 0 则立即重启。
func (app *App) Restart(timeout time.Duration) error {
	if err := app.Shutdown(timeout); err != nil {
		return err
	}

	if err := app.modules.Reset(); err != nil {
		return err
	}

	return app.Run(app.handler)
}
