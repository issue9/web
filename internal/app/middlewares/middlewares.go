// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package middlewares 提供一系列中间
package middlewares

import (
	"net/http"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/host"
	"github.com/issue9/middleware/recovery"

	"github.com/issue9/web/internal/errors"
)

func Handler(h http.Handler, isDebug bool, domains []string, headers map[string]string) http.Handler {
	h = hosts(header(h, headers), domains)
	h = recovery.New(h, errors.Recovery(isDebug))

	// 需保证外层调用不再写入内容。否则可能出错
	h = compress(h, logs.ERROR())

	// NOTE: 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
	if isDebug {
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
