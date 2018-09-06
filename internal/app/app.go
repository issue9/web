// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package app 核心功能的实现
package app

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/issue9/logs"
	"github.com/issue9/middleware"
	"github.com/issue9/mux"

	"github.com/issue9/web/internal/app/middlewares"
	"github.com/issue9/web/internal/app/modules"
	"github.com/issue9/web/internal/app/webconfig"
	"github.com/issue9/web/module"
)

// App 程序运行实例
type App struct {
	webConfig *webconfig.WebConfig

	middleware middleware.Middleware // 应用于全局路由项的中间件
	mux        *mux.Mux
	server     *http.Server

	modules *modules.Modules

	// 当 shutdown 延时关闭时，通过此事件确定 Run() 的返回时机。
	closed chan bool
}

// New 声明一个新的 App 实例
func New(conf *webconfig.WebConfig) (*App, error) {
	app := &App{
		webConfig: conf,
		mux:       mux.New(conf.DisableOptions, false, nil, nil),
		closed:    make(chan bool, 1),
	}

	ms, err := modules.New(50, app.mux, conf)
	if err != nil {
		return nil, err
	}
	app.modules = ms

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

// Modules 获取所有的模块信息
func (app *App) Modules() []*module.Module {
	return app.modules.Modules()
}

// NewModule 声明一个新的模块
func (app *App) NewModule(name, desc string, deps ...string) *module.Module {
	return app.modules.NewModule(name, desc, deps...)
}

// URL 构建一条基于 app.webconfig.URL 的完整 URL
func (app *App) URL(path string) string {
	if len(path) == 0 {
		return app.webConfig.URL
	}

	if path[0] != '/' {
		path = "/" + path
	}
	return app.webConfig.URL + path
}

// InitModules 初始化模块内容
func (app *App) InitModules(tag string, log *log.Logger) error {
	return app.modules.Init(tag, log)
}

// Handler 将当前实例当作一个 http.Handler 返回。一般用于测试。
// 比如在 httptest.NewServer 中使用。
func (app *App) Handler() (http.Handler, error) {
	if app.server == nil {
		if err := app.initServer(); err != nil {
			return nil, err
		}
	}

	return app.server.Handler, nil
}

func (app *App) initServer() error {
	if err := app.InitModules("", nil); err != nil {
		return err
	}

	h := middleware.Handler(app.mux, app.middleware)
	app.server = &http.Server{
		Addr:         ":" + strconv.Itoa(app.webConfig.Port),
		Handler:      middlewares.Handler(h, app.webConfig),
		ErrorLog:     logs.ERROR(),
		ReadTimeout:  app.webConfig.ReadTimeout,
		WriteTimeout: app.webConfig.WriteTimeout,
	}

	return nil
}

// Serve 加载各个模块的数据，运行路由，执行监听程序。
//
// 多次调用，会直接返回 nil 值。
func (app *App) Serve() error {
	// 简单地根据 server 判断是否多次执行。
	// 但是放在多线程环境中同时调用 Serve 可能会出错。
	if app.server != nil {
		return nil
	}

	err := app.initServer()
	if err != nil {
		return err
	}

	conf := app.webConfig
	if !conf.HTTPS {
		err = app.server.ListenAndServe()
	} else {
		err = app.server.ListenAndServeTLS(conf.CertFile, conf.KeyFile)
	}

	// 由 Shutdown() 或 Close() 主动触发的关闭事件，才需要等待其执行完成，
	// 其它错误直接返回，否则一些内部错误会永远卡在此处无法返回。
	if err == http.ErrServerClosed {
		<-app.closed
	}
	return err
}

// Close 关闭服务。
//
// 无论配置文件如果设置，此函数都是直接关闭服务，不会等待。
func (app *App) Close() error {
	logs.Flush()

	if app.server == nil {
		return nil
	}

	app.closed <- true
	return app.server.Close()
}

// Shutdown 关闭所有服务。
//
// 根据配置文件中的配置项，决定当前是直接关闭还是延时之后关闭。
func (app *App) Shutdown() (err error) {
	logs.Flush()

	if app.server == nil {
		return nil
	}

	if app.webConfig.ShutdownTimeout <= 0 {
		app.closed <- true
		return app.server.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), app.webConfig.ShutdownTimeout)
	defer func() {
		cancel()
		app.closed <- true
	}()
	return app.server.Shutdown(ctx)
}
