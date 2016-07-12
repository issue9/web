// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestConfig_init(t *testing.T) {
	a := assert.New(t)
	cfg := &Config{HTTPS: true}

	// 正常加载之后，测试各个变量是否和配置文件中的一样。
	a.NotError(cfg.init())
	a.Equal(":443", cfg.Port).
		Equal(0, len(cfg.Headers)).
		True(cfg.HTTPS)
}

func TestConfig_buildHeaders(t *testing.T) {
	a := assert.New(t)
	fh := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("123"))
	}

	cfg := &Config{Headers: map[string]string{"Server": "test"}}
	h := cfg.buildHeaders(http.HandlerFunc(fh))

	r, err := http.NewRequest("GET", "", nil)
	a.NotError(err).NotNil(r)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Header().Get("Server"), "test")

	// 为空
	cfg = &Config{Headers: map[string]string{}}
	h = cfg.buildHeaders(http.HandlerFunc(fh))

	r, err = http.NewRequest("GET", "", nil)
	a.NotError(err).NotNil(r)

	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Header().Get("Server"), "")
}

func TestConfig_buildPprof(t *testing.T) {
	a := assert.New(t)
	fh := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("123"))
	}

	cfg := &Config{Pprof: "/debug/"}
	h := cfg.buildPprof(http.HandlerFunc(fh))
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/debug/synbol")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	resp, err = http.Get(srv.URL + "/debug/cmdline")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	// 不存在的路由，跳转到fh函数
	resp, err = http.Get(srv.URL + "/debug1/cmdline")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusNotFound)
}
