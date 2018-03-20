// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	stdctx "context"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/host"
	"github.com/issue9/middleware/recovery"

	"github.com/issue9/web/context"
	"github.com/issue9/web/internal/config"
)

const pprofPath = "/debug/pprof/"

// Close 关闭服务。
//
// 无论配置文件如果设置，此函数都是直接关闭服务，不会等待。
func (app *App) Close() error {
	logs.Flush()

	if app.server == nil {
		return nil
	}

	return app.server.Close()
}

// Shutdown 关闭所有服务。
//
// 根据配置文件中的配置项，决定当前是直接关闭还是延时之后关闭。
//
// 同时会使函数 Run() 立即返回 http.ErrServerClosed 错误，
// 请确保该函数不会提前中止程序，致使 Shutdown 失败。
func (app *App) Shutdown() error {
	logs.Flush()

	if app.server == nil {
		return nil
	}

	if app.config.ShutdownTimeout <= 0 {
		return app.server.Close()
	}

	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), app.config.ShutdownTimeout)
	defer cancel()
	return app.server.Shutdown(ctx)
}

func logRecovery(w http.ResponseWriter, msg interface{}) {
	logs.Error(msg)
	context.RenderStatus(w, http.StatusInternalServerError)
}

func buildHandler(conf *config.Config, h http.Handler) http.Handler {
	h = buildHosts(conf, buildHeader(conf, h))

	ff := logRecovery
	if conf.Debug {
		ff = recovery.PrintDebug
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
