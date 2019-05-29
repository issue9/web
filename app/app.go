// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package app 核心功能的实现
package app

import (
	"context"
	"errors"
	"net/http"
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
	http.Server

	uptime        time.Time
	middlewares   *middleware.Manager
	router        *mux.Prefix
	services      []*Service
	scheduled     *scheduled.Server
	logs          *logs.Logs
	webConfig     *webconfig.WebConfig
	errorhandlers *errorhandler.ErrorHandler
	mt            *mimetype.Mimetypes
	compresses    map[string]compress.WriterFunc
	getResult     GetResultFunc
	messages      map[int]*message

	// 当 shutdown 延时关闭时，通过此事件确定 Serve() 的返回时机。
	closed chan struct{}
}

// New 声明一个新的 App 实例
func New(conf *webconfig.WebConfig, get GetResultFunc) *App {
	mux := mux.New(conf.DisableOptions, conf.DisableHead, false, nil, nil)
	middlewares := middleware.NewManager(mux)

	app := &App{
		Server: http.Server{
			Addr:              ":" + strconv.Itoa(conf.Port),
			ErrorLog:          logs.ERROR(),
			ReadTimeout:       conf.ReadTimeout.Duration(),
			WriteTimeout:      conf.WriteTimeout.Duration(),
			IdleTimeout:       conf.IdleTimeout.Duration(),
			ReadHeaderTimeout: conf.ReadHeaderTimeout.Duration(),
			MaxHeaderBytes:    conf.MaxHeaderBytes,
			Handler:           middlewares,
		},
		uptime:        time.Now().In(conf.Location),
		middlewares:   middlewares,
		router:        mux.Prefix(conf.Root),
		services:      make([]*Service, 0, 100),
		scheduled:     scheduled.NewServer(conf.Location),
		logs:          logs.New(),
		webConfig:     conf,
		closed:        make(chan struct{}, 1),
		mt:            mimetype.New(),
		errorhandlers: errorhandler.New(),
		compresses:    make(map[string]compress.WriterFunc, 5),
		getResult:     get,
		messages:      map[int]*message{},
	}

	for url, dir := range conf.Static {
		h := http.StripPrefix(url, http.FileServer(http.Dir(dir)))
		app.router.Get(url+"{path}", h)
	}

	app.AddService(app.scheduledService, "计划任务")

	// 加载固有的中间件，需要在 app 初始化之后调用
	app.buildMiddlewares(conf)

	return app
}

// Uptime 启动的时间
//
// 时区信息与配置文件中的相同
func (app *App) Uptime() time.Time {
	return app.uptime
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

// Run 执行监听程序。
//
// 当调用 Shutdown 关闭服务时，会等待其完成未完的服务，才返回 http.ErrServerClosed
func (app *App) Run() (err error) {
	conf := app.webConfig

	app.runServices()

	if !conf.HTTPS {
		err = app.ListenAndServe()
	} else {
		err = app.ListenAndServeTLS(conf.CertFile, conf.KeyFile)
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
	defer func() {
		app.stopServices()
		app.closed <- struct{}{}
	}()

	return app.Server.Close()
}

// Shutdown 关闭所有服务。
//
// 根据配置文件中的配置项，决定当前是直接关闭还是延时之后关闭。
func (app *App) Shutdown(ctx context.Context) error {
	defer func() {
		app.stopServices()
		app.closed <- struct{}{}
	}()

	err := app.Server.Shutdown(ctx)
	if err != nil && err != context.DeadlineExceeded {
		return err
	}
	return nil
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
