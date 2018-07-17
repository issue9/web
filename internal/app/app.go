// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package app 核心功能的实现
package app

import (
	"net/http"

	"github.com/issue9/middleware"
	"github.com/issue9/mux"

	"github.com/issue9/web/internal/config"
	"github.com/issue9/web/internal/dependency"
	"github.com/issue9/web/module"
)

const configFilename = "web.yaml" // 配置文件的文件名。

// App 程序运行实例
type App struct {
	config    *config.Config
	webConfig *config.Web

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
	conf, err := config.New(configDir)
	if err != nil {
		return nil, err
	}

	app := &App{
		config: conf,
		closed: make(chan bool, 1),
	}

	if err = app.loadConfig(); err != nil {
		return nil, err
	}

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

func (app *App) loadConfig() error {
	conf := &config.Web{}
	err := app.config.Load(app.File(configFilename), conf)
	if err != nil {
		return err
	}

	app.webConfig = conf
	app.mux = mux.New(conf.DisableOptions, false, nil, nil)
	app.router = app.mux.Prefix(conf.Root)

	app.modules = make([]*module.Module, 0, 50)

	return nil
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
	return app.config.File(path...)
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
