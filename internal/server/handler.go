// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/host"
	"github.com/issue9/middleware/recovery"
	"github.com/issue9/middleware/version"
)

func (s *Server) buildHandler(h http.Handler) http.Handler {
	h = s.buildHosts(s.buildVersion(s.buildHeader(h)))

	h = recovery.New(h, func(w http.ResponseWriter, msg interface{}) {
		logs.Error(msg)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	})

	// NOTE: 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
	if s.conf.Pprof {
		h = s.buildPprof(h)
	}

	return h
}

func (s *Server) buildHosts(h http.Handler) http.Handler {
	if len(s.conf.Hosts) == 0 {
		return h
	}

	return host.New(h, s.conf.Hosts...)
}

func (s *Server) buildVersion(h http.Handler) http.Handler {
	if len(s.conf.Version) == 0 {
		return h
	}

	return version.New(h, s.conf.Version, true)
}

func (s *Server) buildHeader(h http.Handler) http.Handler {
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

// 根据 Config.Pprof 决定是否包装调试地址，调用前请确认是否已经开启 Pprof 选项
func (s *Server) buildPprof(h http.Handler) http.Handler {
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
