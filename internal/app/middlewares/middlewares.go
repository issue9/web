// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package middlewares 提供一系列中间
package middlewares

import (
	"net/http"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/middleware/host"
	"github.com/issue9/middleware/recovery"

	"github.com/issue9/web/internal/app/webconfig"
	"github.com/issue9/web/internal/errors"
)

// Handler 将所有配置文件中指定的中间件应用于 h，并返回新的 http.Handler 实例
func Handler(h http.Handler, conf *webconfig.WebConfig) http.Handler {
	h = hosts(header(h, conf.Headers), conf.AllowedDomains)
	h = recovery.New(h, errors.Recovery(conf.Debug))

	// 需保证外层调用不再写入内容。否则可能出错
	if conf.Compress != nil {
		h = compress.New(h, &compress.Options{
			Funcs:    funcs,
			Types:    conf.Compress.Types,
			Size:     conf.Compress.Size,
			ErrorLog: logs.ERROR(),
		})
	}

	// NOTE: 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
	if conf.Debug {
		logs.Debug("调试模式，地址启用：", debugPprofPath, debugVarsPath)
		h = debug(h)
	}

	return h
}

func hosts(h http.Handler, domains []string) http.Handler {
	if len(domains) == 0 {
		return h
	}

	return host.New(h, domains...)
}

func header(h http.Handler, headers map[string]string) http.Handler {
	if len(headers) == 0 {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		h.ServeHTTP(w, r)
	})
}
