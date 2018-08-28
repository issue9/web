// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/web/internal/app/webconfig"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/web/context"
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

	h := Handler(panicFunc, &webconfig.WebConfig{})
	srv := rest.NewServer(t, h, nil)

	// 触发 panic
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusInternalServerError)

	// 触发 panic，调试模式
	h = Handler(panicFunc, &webconfig.WebConfig{Debug: true})
	srv = rest.NewServer(t, h, nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusInternalServerError)

	// 触发 panic, errors.HTTP
	h = Handler(panicHTTPFunc, &webconfig.WebConfig{})
	srv = rest.NewServer(t, h, nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusNotAcceptable)
}

func TestHosts(t *testing.T) {
	a := assert.New(t)
	request := func(h http.Handler, url string, code int) {
		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		a.Equal(w.Code, code)
	}

	// 未指定哉名
	h := hosts(h202, nil)
	request(h, "http://example.com/test", http.StatusAccepted)

	h = hosts(h202, []string{"caixw.io", "example.com"})

	// 带正确的域名访问
	request(h, "http://caixw.io/test", http.StatusAccepted)

	// 带不允许的域名访问
	request(h, "http://not.allowed/test", http.StatusNotFound)
}

func TestHeaders(t *testing.T) {
	h := headers(h202, map[string]string{"Test": "test"})
	srv := rest.NewServer(t, h, nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusAccepted).
		Header("Test", "test")
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
