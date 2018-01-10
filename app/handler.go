// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/host"
	"github.com/issue9/middleware/recovery"
	"github.com/issue9/middleware/version"
	"github.com/issue9/web/context"
)

func logRecovery(w http.ResponseWriter, msg interface{}) {
	logs.Error(msg)
	context.RenderStatus(w, http.StatusInternalServerError)
}

func (s *server) buildHandler(h http.Handler) http.Handler {
	h = s.buildHosts(s.buildVersion(s.buildHeader(h)))

	ff := logRecovery
	if s.conf.Debug {
		ff = recovery.PrintDebug
	}
	h = recovery.New(h, ff)

	// NOTE: 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
	if s.conf.Debug {
		h = s.buildPprof(h)
	}

	return h
}

func (s *server) buildHosts(h http.Handler) http.Handler {
	if len(s.conf.Hosts) == 0 {
		return h
	}

	return host.New(h, s.conf.Hosts...)
}

func (s *server) buildVersion(h http.Handler) http.Handler {
	if len(s.conf.Version) == 0 {
		return h
	}

	return version.New(h, s.conf.Version, true)
}

func (s *server) buildHeader(h http.Handler) http.Handler {
	if len(s.conf.Headers) == 0 {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range s.conf.Headers {
			w.Header().Set(k, v)
		}
		h.ServeHTTP(w, r)
	})
}

// 根据 决定是否包装调试地址，调用前请确认是否已经开启 Pprof 选项
func (s *server) buildPprof(h http.Handler) http.Handler {
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
