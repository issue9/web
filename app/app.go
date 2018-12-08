// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package app 核心功能的实现
package app

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware"
	"github.com/issue9/mux"

	"github.com/issue9/web/config"
	"github.com/issue9/web/internal/modules"
	"github.com/issue9/web/internal/webconfig"
	"github.com/issue9/web/mimetype"
	"github.com/issue9/web/module"
)

const (
	configFilename = "web.yaml" // 配置文件的文件名。
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
)

// App 程序运行实例
type App struct {
	webConfig *webconfig.WebConfig

	middlewares []middleware.Middleware // 应用于全局路由项的中间件
	mux         *mux.Mux
	server      *http.Server

	modules       *modules.Modules
	mt            *mimetype.Mimetypes
	configs       *config.Manager
	logs          *logs.Logs
	errorHandlers map[int]ErrorHandler

	// 当 shutdown 延时关闭时，通过此事件确定 Run() 的返回时机。
	closed chan bool
}

// New 声明一个新的 App 实例
//
// 日志系统会在此处初始化。
// opt 参数在传递之后，再次修改，将不对 App 启作用。
func New(opt *Options) (*App, error) {
	return opt.newApp()
}

// AddMiddlewares 设置全局的中间件，可多次调用。
//
// 在调用 serve 之后调用将不再启效果。
func (app *App) AddMiddlewares(m ...middleware.Middleware) *App {
	app.middlewares = append(app.middlewares, m...)
	return app
}

// Mux 返回 mux.Mux 实例。
func (app *App) Mux() *mux.Mux {
	return app.mux
}

// IsDebug 是否处于调试模式
func (app *App) IsDebug() bool {
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

// InitModules 执行模板的初始化函数。可以重复调用执行。
func (app *App) InitModules(tag string) error {
	return app.modules.Init(tag, app.logs.INFO())
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
	if err := app.InitModules(""); err != nil {
		return err
	}

	app.buildMiddlewares(app.webConfig)
	app.server = &http.Server{
		Addr:              ":" + strconv.Itoa(app.webConfig.Port),
		Handler:           middleware.Handler(app.mux, app.middlewares...),
		ErrorLog:          app.ERROR(),
		ReadTimeout:       app.webConfig.ReadTimeout,
		WriteTimeout:      app.webConfig.WriteTimeout,
		IdleTimeout:       app.webConfig.IdleTimeout,
		ReadHeaderTimeout: app.webConfig.ReadHeaderTimeout,
	}

	return nil
}

// Serve 加载各个模块的数据，运行路由，执行监听程序。
//
// 多次调用，会直接返回 nil 值。
func (app *App) Serve() error {
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
	defer app.logs.Flush()

	if app.server != nil {
		return app.close()
	}
	return nil
}

// Shutdown 关闭所有服务。
//
// 根据配置文件中的配置项，决定当前是直接关闭还是延时之后关闭。
func (app *App) Shutdown() error {
	defer app.logs.Flush()

	if app.server == nil {
		return nil
	}

	if app.webConfig.ShutdownTimeout <= 0 {
		return app.close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), app.webConfig.ShutdownTimeout)
	defer func() {
		cancel()
		app.closed <- true
	}()
	return app.server.Shutdown(ctx)
}

func (app *App) close() error {
	app.closed <- true
	return app.server.Close()
}

// RegisterOnShutdown 等于于 http.Server.RegisterOnShutdown
func (app *App) RegisterOnShutdown(f func()) {
	app.server.RegisterOnShutdown(f)
}

// File 获取文件路径，相对于当前配置目录
func (app *App) File(path string) string {
	return app.configs.File(path)
}

// LoadFile 加载指定的配置文件内容到 v 中
func (app *App) LoadFile(path string, v interface{}) error {
	return app.configs.LoadFile(path, v)
}

// Load 加载指定的配置文件内容到 v 中
func (app *App) Load(r io.Reader, typ string, v interface{}) error {
	return app.configs.Load(r, typ, v)
}

// Mimetypes 返回 mimetype.Mimetypes
func (app *App) Mimetypes() *mimetype.Mimetypes {
	return app.mt
}

// Config 获取 config.Manager 的实例
func (app *App) Config() *config.Manager {
	return app.configs
}

// Grace 指定触发 Shutdown() 的信号，若为空，则任意信号都触发。
//
// 多次调用，则每次指定的信号都会起作用，如果由传递了相同的值，
// 则有可能多次触发 Shutdown()。
//
// NOTE: 传递空值，与不调用，其结果是不同的。
// 若是不调用，则不会处理任何信号；若是传递空值调用，则是处理任何要信号。
func Grace(app *App, sig ...os.Signal) {
	go func() {
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, sig...)

		<-signalChannel
		signal.Stop(signalChannel)

		if err := app.Shutdown(); err != nil {
			app.Error(err)
			app.FlushLogs() // 保证内容会被正常输出到日志。
		}
		close(signalChannel)
	}()
}
