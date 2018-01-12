// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(1)
}

var h1 = http.HandlerFunc(f1)

func TestApp_buildHandler(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)
	app.config = defaultConfig()

	h := app.buildHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("err")
	}))

	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusNotFound)
}

func TestApp_buildHosts_empty(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)
	app.config = defaultConfig()

	h := app.buildHosts(h1)

	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	w := httptest.NewRecorder()
	a.NotNil(r).NotNil(w)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)
}

func TestApp_buildHosts(t *testing.T) {
	a := assert.New(t)
	config := defaultConfig()
	config.Hosts = []string{"caixw.io", "example.com"} // 指定域名
	app := &App{}
	a.NotError(app.initFromConfig(config))

	h := app.buildHosts(h1)

	// 带正确的域名访问
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 带不允许的域名访问
	r = httptest.NewRequest(http.MethodGet, "http://not.exists/test", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusNotFound)
}

func TestApp_buildVersion(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	h := app.buildVersion(h1)

	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 指版本号
	config := defaultConfig()
	config.Version = "1.0"
	app = &App{}
	a.NotError(app.initFromConfig(config))

	h = app.buildVersion(h1)

	// 指版本号的情况下，不正确版本号访问
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	w = httptest.NewRecorder()
	a.NotNil(r).NotNil(w)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusForbidden)

	// 指版本号的情况下，带正确版本号访问
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	r.Header.Set("accept", "application/json;version=1.0")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 指版本号的情况下，带不正确版本号访问
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	r.Header.Set("accept", "application/json;version=2.0")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusNotFound)
}

func TestApp_buildHeader(t *testing.T) {
	a := assert.New(t)
	config := defaultConfig()
	config.Headers = map[string]string{"Test": "test"}
	app := &App{}
	a.NotError(app.initFromConfig(config))

	h := app.buildHeader(h1)

	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)
	a.Equal(w.Header().Get("Test"), "test")
}

func TestApp_buildPprof(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	h := app.buildPprof(h1)

	// 命中 /debug/pprof/cmdline
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/debug/pprof/cmdline", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusOK)

	// 命中 h1
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)
}
