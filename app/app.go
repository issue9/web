// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package app 核心功能的实现
package app

import (
	stdctx "context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware"
	"github.com/issue9/middleware/compress"
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
	compresses    map[string]compress.WriterFunc
	configs       *config.Manager
	logs          *logs.Logs
	errorHandlers map[int]ErrorHandler

	// 当 shutdown 延时关闭时，通过此事件确定 Run() 的返回时机。
	closed chan bool
}

// New 声明一个新的 App 实例
//
// 日志系统会在此处初始化。
func New(conf *Config) (*App, error) {
	mgr := config.NewManager(conf.Dir)
	for k, v := range conf.ConfigUnmarshals {
		if err := mgr.AddUnmarshal(v, k); err != nil {
			return nil, err
		}
	}

	l := logs.New()
	if err := l.InitFromXMLFile(mgr.File(logsFilename)); err != nil {
		return nil, err
	}

	webconf := &webconfig.WebConfig{}
	if err := mgr.LoadFile(configFilename, conf); err != nil {
		return nil, err
	}

	mt := mimetype.New()
	if err := mt.AddMarshals(conf.MimetypeMarshals); err != nil {
		return nil, err
	}
	if err := mt.AddUnmarshals(conf.MimetypeUnmarshals); err != nil {
		return nil, err
	}

	mux := mux.New(webconf.DisableOptions, false, nil, nil)

	ms, err := modules.New(mux, webconf)
	if err != nil {
		return nil, err
	}

	return &App{
		webConfig:     webconf,
		middlewares:   conf.Middlewares,
		mux:           mux,
		closed:        make(chan bool, 1),
		modules:       ms,
		mt:            mt,
		configs:       mgr,
		logs:          l,
		errorHandlers: conf.ErrorHandlers,
		compresses:    conf.Compresses,
	}, nil
}

// SetMiddlewares 设置一个全局的中间件，多次设置，只有最后一次会启作用。
func (app *App) SetMiddlewares(m ...middleware.Middleware) *App {
	app.middlewares = m
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
		ErrorLog:          app.logs.ERROR(),
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

	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), app.webConfig.ShutdownTimeout)
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

// AddMarshals 添加多个编码函数
func (app *App) AddMarshals(ms map[string]mimetype.MarshalFunc) error {
	return app.mt.AddMarshals(ms)
}

// AddMarshal 添加编码函数
func (app *App) AddMarshal(name string, mf mimetype.MarshalFunc) error {
	return app.mt.AddMarshal(name, mf)
}

// AddUnmarshals 添加多个编码函数
func (app *App) AddUnmarshals(ms map[string]mimetype.UnmarshalFunc) error {
	return app.mt.AddUnmarshals(ms)
}

// AddUnmarshal 添加编码函数
func (app *App) AddUnmarshal(name string, mm mimetype.UnmarshalFunc) error {
	return app.mt.AddUnmarshal(name, mm)
}

// AddCompress 添加压缩方法。框架本身已经指定了 gzip 和 deflate 两种方法。
func (app *App) AddCompress(name string, f compress.WriterFunc) error {
	if _, found := app.compresses[name]; found {
		return fmt.Errorf("已经存在同名 %s 的压缩函数", name)
	}

	app.compresses[name] = f
	return nil
}

// SetCompress 修改或是添加压缩方法。
func (app *App) SetCompress(name string, f compress.WriterFunc) {
	app.compresses[name] = f
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

// AddConfig 注册解析函数
func (app *App) AddConfig(m config.UnmarshalFunc, ext ...string) error {
	return app.configs.AddUnmarshal(m, ext...)
}

// SetConfig 修改指定扩展名关联的解析函数，不存在则添加。
func (app *App) SetConfig(m config.UnmarshalFunc, ext ...string) error {
	return app.configs.SetUnmarshal(m, ext...)
}

func (app *App) MimetypeMarshal(name string) (string, mimetype.MarshalFunc, error) {
	return app.mt.Marshal(name)
}

func (app *App) MimetypeUnmarshal(name string) (mimetype.UnmarshalFunc, error) {
	return app.mt.Unmarshal(name)
}
