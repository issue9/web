// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/internal/config"
)

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(1)
}

var h1 = http.HandlerFunc(f1)

func request(a *assert.Assertion, h http.Handler, url string, code int) {
	r := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, code)
}

func TestBuildHandler(t *testing.T) {
	a := assert.New(t)

	panicFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("err")
	})

	// 触发 panic
	conf := &config.Config{}
	h := buildHandler(conf, panicFunc)
	request(a, h, "http://example.com/test", http.StatusInternalServerError)

	// 触发 panic，调试模式
	conf = &config.Config{
		Debug: true,
	}
	h = buildHandler(conf, panicFunc)
	request(a, h, "http://example.com/test", http.StatusNotFound)
}

func TestBuildHosts_empty(t *testing.T) {
	a := assert.New(t)
	conf := &config.Config{}

	h := buildHosts(conf, h1)
	request(a, h, "http://example.com/test", 1)
}

func TestBuildHosts(t *testing.T) {
	a := assert.New(t)
	conf := &config.Config{
		AllowedDomains: []string{"caixw.io", "example.com"},
	}

	h := buildHosts(conf, h1)

	// 带正确的域名访问
	request(a, h, "http://caixw.io/test", 1)

	// 带不允许的域名访问
	request(a, h, "http://not.allowed/test", http.StatusNotFound)
}

func TestBuildHeader(t *testing.T) {
	a := assert.New(t)
	conf := &config.Config{
		Headers: map[string]string{"Test": "test"},
	}

	h := buildHeader(conf, h1)

	r := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)
	a.Equal(w.Header().Get("Test"), "test")
}

func TestBuildPprof(t *testing.T) {
	a := assert.New(t)
	conf := &config.Config{}

	h := buildPprof(conf, h1)

	// 命中 /debug/pprof/cmdline
	request(a, h, "http://example.com/debug/pprof/", http.StatusOK)
	request(a, h, "http://example.com/debug/pprof/cmdline", http.StatusOK)
	request(a, h, "http://example.com/debug/pprof/trace", http.StatusOK)
	request(a, h, "http://example.com/debug/pprof/symbol", http.StatusOK)
	//request(a, h, "http://example.com/debug/pprof/profile", http.StatusOK)

	// 命中 h1
	request(a, h, "http://example.com/debug", 1)
}
