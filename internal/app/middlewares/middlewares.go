// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package middlewares 提供一系列中间
package middlewares

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/middleware/host"
	"github.com/issue9/middleware/recovery"

	"github.com/issue9/web/internal/app/webconfig"
	"github.com/issue9/web/internal/errors"
)

const (
	// 此变量理论上可以更改，但实际上，更改之后，所有的子页面都会不可用。
	debugPprofPath = "/debug/pprof/"

	// 此地址可以修改。
	debugVarsPath = "/debug/vars"
)

var funcs = map[string]compress.WriterFunc{
	"gizp":    compress.NewGzip,
	"deflate": compress.NewDeflate,
}

// AddCompress 添加压缩方法。框架本身已经指定了 gzip 和 deflate 两种方法。
func AddCompress(name string, f compress.WriterFunc) error {
	if _, found := funcs[name]; found {
		return fmt.Errorf("已经存在同名 %s 的压缩函数", name)
	}

	funcs[name] = f
	return nil
}

// SetCompress 修改或是添加压缩方法。
func SetCompress(name string, f compress.WriterFunc) {
	funcs[name] = f
}

// Handler 将所有配置文件中指定的中间件应用于 h，并返回新的 http.Handler 实例
func Handler(h http.Handler, conf *webconfig.WebConfig) http.Handler {
	h = hosts(headers(h, conf.Headers), conf.AllowedDomains)
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

func headers(h http.Handler, headers map[string]string) http.Handler {
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

func debug(h http.Handler) http.Handler {
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
