// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package app 核心功能的实现
package app

import (
	"net/http"
	"path/filepath"

	"github.com/issue9/logs"
	"github.com/issue9/middleware"
	"github.com/issue9/mux"

	"github.com/issue9/web/config"
	"github.com/issue9/web/internal/app/webconfig"
	"github.com/issue9/web/internal/dependency"
	"github.com/issue9/web/module"
)

const (
	configFilename = "web.yaml" // 配置文件的文件名。
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
)

// App 程序运行实例
type App struct {
	configDir string
	webConfig *webconfig.WebConfig

	middleware middleware.Middleware // 应用于全局路由项的中间件
	mux        *mux.Mux
	router     *mux.Prefix
	server     *http.Server

	modules []*module.Module

	// 当 shutdown 延时关闭时，通过此事件确定 Run() 的返回时机。
	closed chan bool
}

// New 声明一个新的 App 实例
func New(configDir string) (*App, error) {
	configDir, err := filepath.Abs(configDir)
	if err != nil {
		return nil, err
	}

	app := &App{
		configDir: configDir,
		closed:    make(chan bool, 1),
	}

	if err = logs.InitFromXMLFile(app.File(logsFilename)); err != nil {
		return nil, err
	}

	conf := &webconfig.WebConfig{}
	if err := app.LoadConfig(configFilename, conf); err != nil {
		return nil, err
	}
	app.webConfig = conf
	app.mux = mux.New(conf.DisableOptions, false, nil, nil)
	app.router = app.mux.Prefix(conf.Root)

	return app, nil
}

// SetMiddleware 设置一个全局的中间件，多次设置，只有最后一次会启作用。
func (app *App) SetMiddleware(m middleware.Middleware) *App {
	app.middleware = m
	return app
}

// Debug 是否处于调试模式
func (app *App) Debug() bool {
	return app.webConfig.Debug
}

// LoadConfig 从配置文件目录加载配置文件到 v 中
func (app *App) LoadConfig(path string, v interface{}) error {
	return config.Load(app.File(path), v)
}

// Modules 获取所有的模块信息
func (app *App) Modules() []*module.Module {
	return app.modules
}

// NewModule 声明一个新的模块
func (app *App) NewModule(name, desc string, deps ...string) *module.Module {
	m := module.New(app.router, name, desc, deps...)
	app.modules = append(app.modules, m)
	return m
}

// File 获取配置目录下的文件名
func (app *App) File(path ...string) string {
	paths := make([]string, 0, len(path)+1)
	paths = append(paths, app.configDir)
	return filepath.Join(append(paths, path...)...)
}

// URL 构建一条基于 app.config.URL 的完整 URL
func (app *App) URL(path string) string {
	if len(path) == 0 {
		return app.webConfig.URL
	}

	if path[0] != '/' {
		path = "/" + path
	}
	return app.webConfig.URL + path
}

// Install 安装各个模块
func (app *App) Install(version string) error {
	dep := dependency.New()
	for _, m := range app.modules {
		dep.Add(m.Name, m.GetInstall(version), m.Deps...)
	}

	return dep.Init()
}
