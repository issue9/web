// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package app 核心功能的实现
package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/middleware/recovery/errorhandler"
	"github.com/issue9/mux/v2"
	"github.com/issue9/scheduled"
	"golang.org/x/text/language"
	xmessage "golang.org/x/text/message"

	"github.com/issue9/web/internal/webconfig"
	"github.com/issue9/web/mimetype"
)

// App 程序运行实例
type App struct {
	middleware.Manager

	modules       []*Module
	router        *mux.Prefix
	services      []*Service
	scheduled     *scheduled.Server
	logs          *logs.Logs
	webConfig     *webconfig.WebConfig
	server        *http.Server
	errorhandlers *errorhandler.ErrorHandler
	mt            *mimetype.Mimetypes
	compresses    map[string]compress.WriterFunc
	getResult     GetResultFunc
	messages      map[int]*message

	// 框架自身用到的模块，除了配置文件中的静态文件服务之外，
	// 还有包含了启动服务等。
	coreModule *Module

	// 当 shutdown 延时关闭时，通过此事件确定 Serve() 的返回时机。
	closed chan struct{}
}

// New 声明一个新的 App 实例
func New(conf *webconfig.WebConfig, logs *logs.Logs, get GetResultFunc) (*App, error) {
	mux := mux.New(conf.DisableOptions, conf.DisableHead, false, nil, nil)

	app := &App{
		Manager:       *middleware.NewManager(mux),
		modules:       make([]*Module, 0, 100),
		router:        mux.Prefix(conf.Root),
		services:      make([]*Service, 0, 100),
		scheduled:     scheduled.NewServer(conf.Location),
		logs:          logs,
		webConfig:     conf,
		closed:        make(chan struct{}, 1),
		mt:            mimetype.New(),
		errorhandlers: errorhandler.New(),
		compresses:    make(map[string]compress.WriterFunc, 5),
		getResult:     get,
		messages:      map[int]*message{},
		server: &http.Server{
			Addr:              ":" + strconv.Itoa(conf.Port),
			ErrorLog:          logs.ERROR(),
			ReadTimeout:       conf.ReadTimeout.Duration(),
			WriteTimeout:      conf.WriteTimeout.Duration(),
			IdleTimeout:       conf.IdleTimeout.Duration(),
			ReadHeaderTimeout: conf.ReadHeaderTimeout.Duration(),
			MaxHeaderBytes:    conf.MaxHeaderBytes,
		},
	}

	app.buildCoreModule(conf)

	if conf.Plugins != "" {
		if err := app.loadPlugins(conf.Plugins); err != nil {
			return nil, err
		}
	}

	app.server.Handler = app

	// 加载固有的中间件，需要在 app 初始化之后调用
	app.buildMiddlewares(conf)

	return app, nil
}

// AddCompresses 添加压缩处理函数
func (app *App) AddCompresses(m map[string]compress.WriterFunc) error {
	for k, v := range m {
		if _, found := app.compresses[k]; found {
			return errors.New("已经存在")
		}

		app.compresses[k] = v
	}

	return nil
}

// Mux 返回相关的 mux.Mux 实例
func (app *App) Mux() *mux.Mux {
	return app.router.Mux()
}

// IsDebug 是否处于调试模式
func (app *App) IsDebug() bool {
	return app.webConfig.Debug
}

// Path 生成路径部分的地址
//
// 基于 app.webConfig.URL 中的路径部分。
func (app *App) Path(p string) string {
	p = app.webConfig.URLPath + p
	if p != "" && p[0] != '/' {
		p = "/" + p
	}

	return p
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

// Serve 加载各个模块的数据，运行路由，执行监听程序。
//
// 当调用 Shutdown 关闭服务时，会等待其完成未完的服务，才返回 http.ErrServerClosed
func (app *App) Serve() (err error) {
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

// LocalPrinter 获取本地化的输出对象
func (app *App) LocalPrinter(tag language.Tag, opts ...xmessage.Option) *xmessage.Printer {
	return xmessage.NewPrinter(tag, opts...)
}

// Close 关闭服务。
//
// 无论配置文件如果设置，此函数都是直接关闭服务，不会等待。
func (app *App) Close() error {
	app.Stop() // 关闭服务

	defer func() {
		app.closed <- struct{}{}
	}()
	return app.server.Close()
}

// Shutdown 关闭所有服务。
//
// 根据配置文件中的配置项，决定当前是直接关闭还是延时之后关闭。
func (app *App) Shutdown() error {
	if app.webConfig.ShutdownTimeout <= 0 {
		return app.Close()
	}

	app.Stop() // 关闭服务
	dur := app.webConfig.ShutdownTimeout.Duration()
	ctx, cancel := context.WithTimeout(context.Background(), dur)

	defer func() {
		cancel()
		app.closed <- struct{}{}
	}()
	err := app.server.Shutdown(ctx)
	if err != nil && err != context.DeadlineExceeded {
		return err
	}
	return nil
}

// Server 获取 http.Server 实例
func (app *App) Server() *http.Server {
	return app.server
}

// Mimetypes 返回 mimetype.Mimetypes
func (app *App) Mimetypes() *mimetype.Mimetypes {
	return app.mt
}

// ErrorHandlers 错误处理功能
func (app *App) ErrorHandlers() *errorhandler.ErrorHandler {
	return app.errorhandlers
}

// Location 当前设置的时区信息
func (app *App) Location() *time.Location {
	return app.webConfig.Location
}

// Logs 返回 logs.Logs 实例
func (app *App) Logs() *logs.Logs {
	return app.logs
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
			app.Logs().Error(err)
		}
		app.Logs().Flush() // 保证内容会被正常输出到日志。
		close(signalChannel)
	}()
}
