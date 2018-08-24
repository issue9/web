// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	stdctx "context"
	"expvar"
	"net/http"
	"net/http/pprof"
	"path"
	"strconv"
	"strings"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/middleware/host"
	"github.com/issue9/middleware/recovery"

	"github.com/issue9/web/errorhandler"
	"github.com/issue9/web/internal/dependency"
)

// 定义了用于调试输出的网页地址。
const (
	// 此变量理论上可以更改，但实际上，更改之后，所有的子页面都会不可用。
	debugPprofPath = "/debug/pprof/"

	// 此地址可以修改。
	debugVarsPath = "/debug/vars"
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
		Handler:      buildHandler(app.webConfig, h),
		ErrorLog:     logs.ERROR(),
		ReadTimeout:  app.webConfig.ReadTimeout,
		WriteTimeout: app.webConfig.WriteTimeout,
	}

	return nil
}

func (app *App) initModules() error {
	// 在初始化模块之前，先加载插件
	if app.webConfig.Plugins != "" {
		if err := app.loadPlugins(app.webConfig.Plugins); err != nil {
			return err
		}
	}

	dep := dependency.New()

	for _, module := range app.modules {
		if err := dep.Add(module.Name, module.GetInit(), module.Deps...); err != nil {
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

func buildHandler(conf *webconfig, h http.Handler) http.Handler {
	h = buildHosts(conf, buildHeader(conf, h))
	h = recovery.New(h, errorhandler.Recovery(conf.Debug))

	// 需保证外层调用不再写入内容。否则可能出错
	h = compress.New(h, logs.ERROR(), map[string]compress.BuildCompressWriter{
		"gizp":    compress.NewGzip,
		"deflate": compress.NewDeflate,
	})

	// NOTE: 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
	if conf.Debug {
		logs.Debug("调试模式，地址启用：", debugPprofPath, debugVarsPath)
		h = buildDebug(h)
	}

	return h
}

func buildHosts(conf *webconfig, h http.Handler) http.Handler {
	if len(conf.AllowedDomains) == 0 {
		return h
	}

	return host.New(h, conf.AllowedDomains...)
}

func buildHeader(conf *webconfig, h http.Handler) http.Handler {
	if len(conf.Headers) == 0 {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range conf.Headers {
			w.Header().Set(k, v)
		}
		h.ServeHTTP(w, r)
	})
}

// 根据 决定是否包装调试地址，调用前请确认是否已经开启 Pprof 选项
func buildDebug(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, debugPprofPath):
			path := r.URL.Path[len(debugPprofPath):]
			switch path {
			case "cmdline":
				pprof.Cmdline(w, r)
			case "profile":
				pprof.Profile(w, r)
			case "symbol":
				pprof.Symbol(w, r)
			case "trace":
				pprof.Trace(w, r)
			default:
				pprof.Index(w, r)
			}
		case strings.HasPrefix(r.URL.Path, debugVarsPath):
			expvar.Handler().ServeHTTP(w, r)
		default:
			h.ServeHTTP(w, r)
		}
	}) // end return http.HandlerFunc
}
