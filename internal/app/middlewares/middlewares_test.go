// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package middlewares

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/rest"
	"github.com/issue9/middleware"

	"github.com/issue9/web/context"
	"github.com/issue9/web/internal/app/webconfig"
)

var h202 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
})

func TestHandler(t *testing.T) {
	panicFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("err")
	})

	panicHTTPFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context.Exit(http.StatusNotAcceptable)
	})

	h := middleware.Handler(panicFunc, Middlewares(&webconfig.WebConfig{})...)
	srv := rest.NewServer(t, h, nil)

	// 触发 panic
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusInternalServerError)

	// 触发 panic，调试模式
	h = middleware.Handler(panicFunc, Middlewares(&webconfig.WebConfig{Debug: true})...)
	srv = rest.NewServer(t, h, nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusInternalServerError)

	// 触发 panic, errors.HTTP
	h = middleware.Handler(panicHTTPFunc, Middlewares(&webconfig.WebConfig{})...)
	srv = rest.NewServer(t, h, nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusNotAcceptable)
}

func TestDebug(t *testing.T) {
	srv := rest.NewServer(t, debug(h202), nil)

	// 命中 /debug/pprof/cmdline
	srv.NewRequest(http.MethodGet, "/debug/pprof/").
		Do().
		Status(http.StatusOK)

	srv.NewRequest(http.MethodGet, "/debug/pprof/cmdline").
		Do().
		Status(http.StatusOK)

	srv.NewRequest(http.MethodGet, "/debug/pprof/trace").
		Do().
		Status(http.StatusOK)

	srv.NewRequest(http.MethodGet, "/debug/pprof/symbol").
		Do().
		Status(http.StatusOK)

	// /debug/vars
	srv.NewRequest(http.MethodGet, "/debug/vars").
		Do().
		Status(http.StatusOK)

	// 命中 h202
	srv.NewRequest(http.MethodGet, "/debug/").
		Do().
		Status(http.StatusAccepted)
}
