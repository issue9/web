// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 一个模块化的微形web框架。
//  m, err := web.NewModule("m1")
//  m.Get("/", ...).
//    Post("/", ...)
//  // 其它模块的初始化工作...
//  web.Run(&Config{},errhandler) // 开始监听端口
package web

import (
	"net/http"
	"net/http/pprof"
	"path/filepath"
	"strings"

	"github.com/issue9/context"
	"github.com/issue9/mux"
	"github.com/issue9/term/colors"
)

// web包的相关配置内容。
type Config struct {
	HTTPS      bool              `json:"https"`                // 是否启用https
	CertFile   string            `json:"certFile,omitempty"`   // 当https为true时，此值为必须
	KeyFile    string            `json:"keyFile,omitempty"`    // 当https为true时，此值为必须
	Port       string            `json:"port,omitempty"`       // 端口，不指定，默认为80或是443
	ServerName string            `json:"serverName,omitempty"` // 响应头的server变量，为空时，不输出该内容
	Static     map[string]string `json:"static,omitempty"`     // 静态路由映身，键名表示路由路径，键值表示文件目录
	Pprof      string            `json:"pprof,omitempty"`      // 指定pprof地址
	ErrHandler mux.RecoverFunc   `json:"-"`                    // 错误处理
}

// 检测cfg的各项字段是否合法，
func checkConfig(cfg *Config) {
	// Port检测
	if len(cfg.Port) == 0 {
		if cfg.HTTPS {
			cfg.Port = ":443"
		} else {
			cfg.Port = ":80"
		}
	}
	if cfg.Port[0] != ':' {
		cfg.Port = ":" + cfg.Port
	}

	// 确保每个目录都以/结尾
	for k, v := range cfg.Static {
		last := v[len(v)-1]
		if last != filepath.Separator && last != '/' {
			v += string(filepath.Separator)
		}
		cfg.Static[k] = v
	}
}

// 初始化web包的内容。
func Run(cfg *Config) {
	checkConfig(cfg)

	if len(cfg.Static) > 0 {
		group, err := NewModule("static")
		if err != nil {
			panic(err)
		}

		for url, dir := range cfg.Static {
			group.Get(url, http.StripPrefix(url, http.FileServer(http.Dir(dir))))
		}
	}

	listen(cfg)
}

// 开始监听。
// errorHandler 为错误处理函数。
func listen(cfg *Config) {
	var h http.Handler = serveMux

	// 修改服务器名称
	if len(cfg.ServerName) > 0 {
		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Server", cfg.ServerName)
			h.ServeHTTP(w, r)
		})
	}

	// 作一些清理和错误处理
	h = mux.NewRecovery(context.FreeHandler(serveMux), cfg.ErrHandler)

	// 在最外层添加调试地址，保证调试内容不会被其它handler干扰。
	ref := h
	if len(cfg.Pprof) > 0 {
		colors.Println(colors.Stdout, colors.Green, colors.Default, "开启了调试功能，地址为：", cfg.Pprof)

		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, cfg.Pprof) {
				ref.ServeHTTP(w, r)
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

	if h == nil {
		panic("h==nil")
	}

	if cfg.HTTPS {
		http.ListenAndServeTLS(cfg.Port, cfg.CertFile, cfg.KeyFile, h)
	} else {
		http.ListenAndServe(cfg.Port, h)
	}
}
