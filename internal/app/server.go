// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	stdctx "context"
	"net/http"
	"path"
	"strconv"

	"github.com/issue9/logs"

	"github.com/issue9/web/internal/app/middlewares"
	"github.com/issue9/web/internal/app/plugin"
	"github.com/issue9/web/internal/dependency"
)

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
	for url, dir := range app.webConfig.Static {
		pattern := path.Join(app.webConfig.Root, url+"{path}")
		fs := http.FileServer(http.Dir(dir))
		app.router.Get(pattern, http.StripPrefix(url, fs))
	}

	if err := app.initModules(); err != nil {
		return err
	}

	var h http.Handler = app.mux
	if app.middleware != nil {
		h = app.middleware(app.mux)
	}

	app.server = &http.Server{
		Addr:         ":" + strconv.Itoa(app.webConfig.Port),
		Handler:      middlewares.Handler(h, app.webConfig),
		ErrorLog:     logs.ERROR(),
		ReadTimeout:  app.webConfig.ReadTimeout,
		WriteTimeout: app.webConfig.WriteTimeout,
	}

	return nil
}

func (app *App) initModules() (err error) {
	// 在初始化模块之前，先加载插件
	app.modules, err = plugin.Load(app.webConfig.Plugins, app.router)
	if err != nil {
		return err
	}

	dep := dependency.New()
	for _, module := range app.modules {
		if err = dep.Add(module.Name, module.GetInit(), module.Deps...); err != nil {
			return err
		}
	}
	return dep.Init()
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

	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), app.webConfig.ShutdownTimeout)
	defer func() {
		cancel()
		app.closed <- true
	}()
	return app.server.Shutdown(ctx)
}
