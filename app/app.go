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

	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/middleware/recovery/errorhandler"
	"github.com/issue9/mux/v2"
	"golang.org/x/text/message"

	"github.com/issue9/web/config"
	"github.com/issue9/web/internal/messages"
	"github.com/issue9/web/internal/modules"
	"github.com/issue9/web/internal/webconfig"
	"github.com/issue9/web/mimetype"
	"github.com/issue9/web/module"
)

// 框加需要用到的配置文件名。
// 实际路径需要通过 App.File 获取。
const (
	ConfigFilename = "web.yaml"
	LogsFilename   = "logs.xml"
)

// App 程序运行实例
type App struct {
	webConfig *webconfig.WebConfig

	server *http.Server

	modules       *modules.Modules
	configs       *config.Manager
	logs          *logs.Logs
	errorhandlers *errorhandler.ErrorHandler
	mt            *mimetype.Mimetypes
	messages      *messages.Messages
	compresses    map[string]compress.WriterFunc

	// 当 shutdown 延时关闭时，通过此事件确定 Serve() 的返回时机。
	closed chan bool
}

// New 声明一个新的 App 实例
//
// 日志系统会在此处初始化。
func New(mgr *config.Manager) (*App, error) {
	logs := logs.New()
	if err := logs.InitFromXMLFile(mgr.File(LogsFilename)); err != nil {
		return nil, err
	}

	webconf := &webconfig.WebConfig{}
	if err := mgr.LoadFile(ConfigFilename, webconf); err != nil {
		if serr, ok := err.(*config.Error); ok {
			serr.File = LogsFilename
		}
		return nil, err
	}

	ms, err := modules.New(webconf)
	if err != nil {
		return nil, err
	}

	app := &App{
		webConfig:     webconf,
		closed:        make(chan bool, 1),
		modules:       ms,
		mt:            mimetype.New(),
		configs:       mgr,
		logs:          logs,
		errorhandlers: errorhandler.New(),
		messages:      messages.New(),
		compresses:    make(map[string]compress.WriterFunc, 5),
		server: &http.Server{
			Addr:              ":" + strconv.Itoa(webconf.Port),
			ErrorLog:          logs.ERROR(),
			ReadTimeout:       webconf.ReadTimeout,
			WriteTimeout:      webconf.WriteTimeout,
			IdleTimeout:       webconf.IdleTimeout,
			ReadHeaderTimeout: webconf.ReadHeaderTimeout,
			MaxHeaderBytes:    webconf.MaxHeaderBytes,
		},
	}
	app.server.Handler = app

	// 加载固有的中间件，需要在 ms 初始化之后调用
	app.buildMiddlewares(webconf)

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

// AddMiddlewares 设置全局的中间件，可多次调用。
func (app *App) AddMiddlewares(m middleware.Middleware) *App {
	app.modules.After(m)
	return app
}

// Mux 返回 mux.Mux 实例。
func (app *App) Mux() *mux.Mux {
	return app.modules.Mux()
}

// ErrorHandlers 错误处理功能
func (app *App) ErrorHandlers() *errorhandler.ErrorHandler {
	return app.errorhandlers
}

// IsDebug 是否处于调试模式
func (app *App) IsDebug() bool {
	return app.webConfig.Debug
}

// Modules 获取所有的模块信息
func (app *App) Modules() []*module.Module {
	return app.modules.Modules()
}

// Tags 获取所有的子模块名称
func (app *App) Tags() []string {
	return app.modules.Tags()
}

// NewModule 声明一个新的模块
func (app *App) NewModule(name, desc string, deps ...string) *module.Module {
	return app.modules.NewModule(name, desc, deps...)
}

// Path 获取中径部分的地址
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

// InitModules 执行模板的初始化函数。可以重复调用执行。
func (app *App) InitModules(tag string) error {
	return app.modules.Init(tag, app.Logs().INFO())
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.modules.ServeHTTP(w, r)
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

// Close 关闭服务。
//
// 无论配置文件如果设置，此函数都是直接关闭服务，不会等待。
func (app *App) Close() error {
	defer app.Logs().Flush()

	return app.close()
}

// Shutdown 关闭所有服务。
//
// 根据配置文件中的配置项，决定当前是直接关闭还是延时之后关闭。
func (app *App) Shutdown() error {
	defer app.Logs().Flush()

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

// Server 获取 http.Server 实例
func (app *App) Server() *http.Server {
	return app.server
}

// Mimetypes 返回 mimetype.Mimetypes
func (app *App) Mimetypes() *mimetype.Mimetypes {
	return app.mt
}

// Config 获取 config.Manager 的实例
func (app *App) Config() *config.Manager {
	return app.configs
}

// NewMessages 添加新的错误消息
func (app *App) NewMessages(status int, msgs map[int]string) {
	app.messages.NewMessages(status, msgs)
}

// Messages 获取所有的错误消息代码
//
// 如果指定 p 的值，则返回本地化的消息内容。
func (app *App) Messages(p *message.Printer) map[int]string {
	if p == nil {
		return app.messages.Messages()
	}

	return app.messages.LocaleMessages(p)
}

// Message 查找指定代码的错误信息
func (app *App) Message(code int) (*messages.Message, bool) {
	return app.messages.Message(code)
}

// Logs 获取 logs.Logs 实例
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
			app.Logs().Flush() // 保证内容会被正常输出到日志。
		}
		close(signalChannel)
	}()
}
