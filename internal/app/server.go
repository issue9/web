// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	stdctx "context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"path"
	"strconv"
	"strings"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/middleware/host"
	"github.com/issue9/middleware/recovery"

	"github.com/issue9/web/context"
	"github.com/issue9/web/errors"
	"github.com/issue9/web/internal/config"
)

const pprofPath = "/debug/pprof/"

// Run 加载各个模块的数据，运行路由，执行监听程序。
func (app *App) Run() (err error) {
	if err = app.modules.Init(); err != nil {
		return err
	}

	// 静态文件路由，在其它路由构建之前调用
	for url, dir := range app.config.Static {
		pattern := path.Join(app.config.Root, url+"{path}")
		fs := http.FileServer(http.Dir(dir))
		app.router.Get(pattern, http.StripPrefix(url, compress.New(fs, logs.ERROR())))
	}

	var h http.Handler = app.mux
	if app.middleware != nil {
		h = app.middleware(app.mux)
	}

	app.server = &http.Server{
		Addr:         ":" + strconv.Itoa(app.config.Port),
		Handler:      buildHandler(app.config, h),
		ErrorLog:     logs.ERROR(),
		ReadTimeout:  app.config.ReadTimeout,
		WriteTimeout: app.config.WriteTimeout,
	}

	if !app.config.HTTPS {
		err = app.server.ListenAndServe()
	} else {
		err = app.server.ListenAndServeTLS(app.config.CertFile, app.config.KeyFile)
	}

	// 由 Shutdown 或 Close() 主动触发的关闭事件，才需要等待其执行完成，
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

	if app.config.ShutdownTimeout <= 0 {
		app.closed <- true
		return app.server.Close()
	}

	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), app.config.ShutdownTimeout)
	defer func() {
		cancel()
		app.closed <- true
	}()
	return app.server.Shutdown(ctx)
}

func logRecovery(w http.ResponseWriter, msg interface{}) {
	fmt.Println("logs====")
	logs.Error(msg)

	if err, ok := msg.(errors.HTTP); ok {
		context.RenderStatus(w, int(err))
		return
	}

	context.RenderStatus(w, http.StatusInternalServerError)
}

func debugRecovery(w http.ResponseWriter, msg interface{}) {
	fmt.Println("debug====")
	if err, ok := msg.(errors.HTTP); ok {
		context.RenderStatus(w, int(err))
		w.Write([]byte(errors.TraceStack(3, err.Error())))
		return
	}

	errors.TraceStack(3, msg)
	context.RenderStatus(w, http.StatusInternalServerError)
}

func buildHandler(conf *config.Config, h http.Handler) http.Handler {
	h = buildHosts(conf, buildHeader(conf, h))

	ff := logRecovery
	if conf.Debug {
		ff = debugRecovery
	}
	h = recovery.New(h, ff)

	// NOTE: 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
	if conf.Debug {
		h = buildPprof(h)
	}

	return h
}

func buildHosts(conf *config.Config, h http.Handler) http.Handler {
	if len(conf.AllowedDomains) == 0 {
		return h
	}

	return host.New(h, conf.AllowedDomains...)
}

func buildHeader(conf *config.Config, h http.Handler) http.Handler {
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
func buildPprof(h http.Handler) http.Handler {
	logs.Debug("开启了调试功能，地址为：", pprofPath)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, pprofPath) {
			h.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path[len(pprofPath):]
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
	}) // end return http.HandlerFunc
}
