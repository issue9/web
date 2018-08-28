// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

	h := Handler(panicFunc, false, nil, nil)
	srv := rest.NewServer(t, h, nil)

	// 触发 panic
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusInternalServerError)

	// 触发 panic，调试模式
	h = Handler(panicFunc, true, nil, nil)
	srv = rest.NewServer(t, h, nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusInternalServerError)

	// 触发 panic, errors.HTTP
	h = Handler(panicHTTPFunc, false, nil, nil)
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

func TestBuildHeader(t *testing.T) {
	h := header(h202, map[string]string{"Test": "test"})
	srv := rest.NewServer(t, h, nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusAccepted).
		Header("Test", "test")
}
