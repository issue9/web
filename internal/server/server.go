// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package server 根据配置内容，生成相应的服务监听程序。
package server

import (
	stdctx "context"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/host"
	"github.com/issue9/middleware/recovery"

	"github.com/issue9/web/context"
	"github.com/issue9/web/internal/config"
)

const pprofPath = "/debug/pprof/"

var (
	conf   *config.Config
	server *http.Server
)

// Listen 监听程序
func Listen(h http.Handler, config *config.Config) error {
	conf = config

	server = &http.Server{
		Addr:         ":" + strconv.Itoa(conf.Port),
		Handler:      buildHandler(h), // 依赖全局变量 conf
		ErrorLog:     logs.ERROR(),
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	}

	if !conf.HTTPS {
		return server.ListenAndServe()
	}

	return server.ListenAndServeTLS(conf.CertFile, conf.KeyFile)
}

// Close 立即关闭服务
func Close() error {
	logs.Flush()

	if server == nil {
		return nil
	}

	return server.Close()
}

// Shutdown 关闭所有服务。
//
// 和 Close 的区别在于 Shutdown 会等待所有的服务完成之后才关闭，
// 等待时间由 conf.shutdownTimeout 决定。
func Shutdown() error {
	logs.Flush()

	if server == nil {
		return nil
	}

	if conf == nil || conf.ShutdownTimeout <= 0 {
		return server.Close()
	}

	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), conf.ShutdownTimeout)
	defer cancel()
	return server.Shutdown(ctx)
}

func logRecovery(w http.ResponseWriter, msg interface{}) {
	logs.Error(msg)
	context.RenderStatus(w, http.StatusInternalServerError)
}

func buildHandler(h http.Handler) http.Handler {
	h = buildHosts(buildHeader(h))

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

func buildHosts(h http.Handler) http.Handler {
	if len(conf.AllowedDomains) == 0 {
		return h
	}

	return host.New(h, conf.AllowedDomains...)
}

func buildHeader(h http.Handler) http.Handler {
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
