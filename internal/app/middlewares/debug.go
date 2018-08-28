// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package middlewares

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"strings"
)

const (
	// 此变量理论上可以更改，但实际上，更改之后，所有的子页面都会不可用。
	debugPprofPath = "/debug/pprof/"

	// 此地址可以修改。
	debugVarsPath = "/debug/vars"
)

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
