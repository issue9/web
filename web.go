// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 一个模块化的微形web框架。
//  m, err := web.NewModule("m1")
//  m.Get("/", ...).
//    Post("/", ...)
//  // 其它模块的初始化工作...
//  web.Run(&Config{}) // 开始监听端口
//
// web依赖logs包作日志输出，请确保已经正确初始化该包。
package web

import (
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/issue9/context"
	"github.com/issue9/handlers"
	"github.com/issue9/logs"
)

// 启动Run()函数的相关参数。
type Config struct {
	HTTPS      bool                 `json:"https,omitempty"`    // 是否启用https
	CertFile   string               `json:"certFile,omitempty"` // 当https为true时，此值为必填
	KeyFile    string               `json:"keyFile,omitempty"`  // 当https为true时，此值为必填
	Port       string               `json:"port,omitempty"`     // 端口，不指定，默认为80或是443
	Headers    map[string]string    `json:"headers,omitempty"`  // 附加的头信息，头信息可能在其它地方被修改
	Pprof      string               `json:"pprof,omitempty"`    // 指定pprof地址
	ErrHandler handlers.RecoverFunc `json:"-"`                  // 错误处理
}

// 检测cfg的各项字段是否合法，
func (cfg *Config) init() {
	if len(cfg.Port) == 0 {
		if cfg.HTTPS {
			cfg.Port = ":443"
		} else {
			cfg.Port = ":80"
		}
	} else if cfg.Port[0] != ':' {
		cfg.Port = ":" + cfg.Port
	}
}

// 修改服务器名称
func (cfg *Config) buildHeaders(h http.Handler) http.Handler {
	if len(cfg.Headers) > 0 {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for k, v := range cfg.Headers {
				w.Header().Set(k, v)
			}
			h.ServeHTTP(w, r)
		})
	}

	return h
}

// 根据config.Pprof决定是否包装调试地址
func (cfg *Config) buildPprof(h http.Handler) http.Handler {
	if len(cfg.Pprof) > 0 {
		logs.Debug("web:", "开启了调试功能，地址为：", cfg.Pprof)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, cfg.Pprof) {
				h.ServeHTTP(w, r)
				return
			}

			path := r.URL.Path[len(cfg.Pprof):]
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
		})
	}

	return h
}

// 初始化web包的内容。
func Run(cfg *Config) {
	cfg.init()

	listen(cfg)
}

// 开始监听。
func listen(cfg *Config) {
	h := cfg.buildHeaders(serveMux)

	// 作一些清理和错误处理
	h = handlers.Recovery(context.FreeHandler(h), cfg.ErrHandler)

	// 在最外层添加调试地址，保证调试内容不会被其它handler干扰。
	h = cfg.buildPprof(h)

	if cfg.HTTPS {
		http.ListenAndServeTLS(cfg.Port, cfg.CertFile, cfg.KeyFile, h)
	} else {
		http.ListenAndServe(cfg.Port, h)
	}
}
