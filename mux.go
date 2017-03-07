// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/issue9/context"
	"github.com/issue9/handlers"
	"github.com/issue9/logs"
	"github.com/issue9/mux"
	"github.com/issue9/web/internal/config"
)

// 路由控制器
var defaultServeMux = mux.NewServeMux()

// Mux 返回默认的 *mux.ServeMux 实例
func Mux() *mux.ServeMux {
	return defaultServeMux
}

// Run 运行路由，执行监听程序。
func run(conf *config.Config, mux *mux.ServeMux) error {
	// 在其它之前调用
	if err := buildStaticModule(conf, mux); err != nil {
		return err
	}

	h := buildHandler(conf, mux)

	if conf.HTTPS {
		switch conf.HTTPState {
		case config.HTTPStateListen:
			logs.Infof("开始监听%v端口", config.HTTPPort)
			go getServer(conf, config.HTTPPort, h).ListenAndServe()
		case config.HTTPStateRedirect:
			logs.Infof("开始监听%v端口", config.HTTPPort)
			go httpRedirectListenAndServe(conf)
			// 空值或是 disable 均为默认处理方式
		}

		logs.Infof("开始监听%v端口", conf.Port)
		return getServer(conf, conf.Port, h).ListenAndServeTLS(conf.CertFile, conf.KeyFile)
	}

	logs.Infof("开始监听%v端口", conf.Port)
	return getServer(conf, conf.Port, h).ListenAndServe()
}

// 构建一个静态文件服务模块
func buildStaticModule(conf *config.Config, mux *mux.ServeMux) error {
	if len(conf.Static) == 0 {
		return nil
	}

	for url, dir := range conf.Static {
		if !strings.HasSuffix(url, "/") {
			url += "/"
		}
		mux.Get(url, http.StripPrefix(url, handlers.Compress(http.FileServer(http.Dir(dir)))))
	}

	return nil
}

// 构建一个从 HTTP 跳转到 HTTPS 的路由服务。
func httpRedirectListenAndServe(conf *config.Config) error {
	srv := getServer(conf, config.HTTPPort, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL
		url.Host = r.Host + conf.Port
		url.Scheme = "HTTPS"

		urlStr := url.String()
		http.Redirect(w, r, urlStr, http.StatusMovedPermanently)
	}))

	return srv.ListenAndServe()
}

func buildHandler(conf *config.Config, h http.Handler) http.Handler {
	h = buildHeader(conf, h)

	// 清理 context 的相关内容
	h = context.FreeHandler(h)

	// 若是调试状态，则向客户端输出详细错误信息
	if len(conf.Pprof) > 0 {
		h = handlers.Recovery(h, handlers.PrintDebug)
		// NOTE: 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
		return buildPprof(conf, h)
	}

	return h
}

// 根据 Config.Pprof 决定是否包装调试地址，调用前请确认是否已经开启 Pprof 选项
func buildPprof(conf *config.Config, h http.Handler) http.Handler {
	logs.Debug("web:", "开启了调试功能，地址为：", conf.Pprof)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, conf.Pprof) {
			h.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path[len(conf.Pprof):]
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

// 获取 http.Server 实例，相对于 http 的默认实现，指定了 ErrorLog 字段。
func getServer(conf *config.Config, port string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:         port,
		Handler:      h,
		ErrorLog:     logs.ERROR(),
		ReadTimeout:  conf.ReadTimeout * time.Second,
		WriteTimeout: conf.WriteTimeout * time.Second,
	}
}
