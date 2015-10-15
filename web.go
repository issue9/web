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
	"path/filepath"

	"github.com/issue9/context"
	"github.com/issue9/mux"
)

// web包的相关配置内容。
type Config struct {
	HTTPS      bool              `json:"https"`            // 是否启用https
	CertFile   string            `json:"certFile"`         // 当https为true时，此值为必须
	KeyFile    string            `json:"keyFile"`          // 当https为true时，此值为必须
	Port       string            `json:"port"`             // 端口，不指定，默认为80或是443
	ServerName string            `json:"serverName"`       // 响应头的server变量，为空时，不输出该内容
	Static     map[string]string `json:"static,omitempty"` // 静态路由映身，键名表示路由路径，键值表示文件目录
	ErrHandler mux.RecoverFunc   `json:"-"`                // 错误处理
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

// 开始监听。
// errorHandler 为错误处理函数。
func listen(cfg *Config) {
	hf := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if len(cfg.ServerName) > 0 {
			w.Header().Add("Server", cfg.ServerName) // 添加serverName
		}

		serveMux.ServeHTTP(w, req)
	})

	h := context.FreeHandler(hf)
	if cfg.HTTPS {
		http.ListenAndServeTLS(cfg.Port, cfg.CertFile, cfg.KeyFile, mux.NewRecovery(h, cfg.ErrHandler))
	} else {
		http.ListenAndServe(cfg.Port, mux.NewRecovery(h, cfg.ErrHandler))
	}
}
