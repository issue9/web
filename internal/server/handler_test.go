// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

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

func TestServer_buildHandler(t *testing.T) {
	a := assert.New(t)

	s, err := New(DefaultConfig())
	a.NotError(err).NotNil(s)
	h := s.buildHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("err")
	}))

	r := httptest.NewRequest("GET", "http://caixw.io/test", nil)
	w := httptest.NewRecorder()
	a.NotNil(r).NotNil(w)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusInternalServerError)
}

func TestServer_buildHosts_empty(t *testing.T) {
	a := assert.New(t)

	s, err := New(DefaultConfig())
	a.NotError(err).NotNil(s)
	h := s.buildHosts(h1)

	r := httptest.NewRequest("GET", "http://caixw.io/test", nil)
	w := httptest.NewRecorder()
	a.NotNil(r).NotNil(w)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)
}

func TestServer_buildHosts(t *testing.T) {
	a := assert.New(t)

	// 指定域名
	c := DefaultConfig()
	c.Hosts = []string{"caixw.io", "example.com"}
	s, err := New(c)
	a.NotError(err).NotNil(s)
	h := s.buildHosts(h1)

	// 带正确的域名访问
	r := httptest.NewRequest("GET", "http://caixw.io/test", nil)
	w := httptest.NewRecorder()
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 带不允许的域名访问
	r = httptest.NewRequest("GET", "http://not.exists/test", nil)
	w = httptest.NewRecorder()
	a.NotNil(r).NotNil(w)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusForbidden)
}

func TestServer_buildVersion(t *testing.T) {
	a := assert.New(t)

	s, err := New(DefaultConfig())
	a.NotError(err).NotNil(s)
	h := s.buildVersion(h1)

	r := httptest.NewRequest("GET", "http://caixw.io/test", nil)
	w := httptest.NewRecorder()
	a.NotNil(r).NotNil(w)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 指版本号
	c := DefaultConfig()
	c.Version = "1.0"
	s, err = New(c)
	a.NotError(err).NotNil(s)
	h = s.buildVersion(h1)

	// 指版本号的情况下，不正确版本号访问
	r = httptest.NewRequest("GET", "http://caixw.io/test", nil)
	w = httptest.NewRecorder()
	a.NotNil(r).NotNil(w)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusForbidden)

	// 指版本号的情况下，带正确版本号访问
	r = httptest.NewRequest("GET", "http://caixw.io/test", nil)
	r.Header.Set("accept", "application/json;version=1.0")
	w = httptest.NewRecorder()
	a.NotNil(r).NotNil(w)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 指版本号的情况下，带不正确版本号访问
	r = httptest.NewRequest("GET", "http://caixw.io/test", nil)
	r.Header.Set("accept", "application/json;version=2.0")
	w = httptest.NewRecorder()
	a.NotNil(r).NotNil(w)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusForbidden)
}

func TestServer_buildHeader(t *testing.T) {
	a := assert.New(t)

	c := DefaultConfig()
	c.Headers = map[string]string{"Test": "test"}
	s, err := New(c)
	a.NotError(err).NotNil(s)
	h := s.buildHeader(h1)

	r := httptest.NewRequest("GET", "http://caixw.io/test", nil)
	w := httptest.NewRecorder()
	a.NotNil(r).NotNil(w)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)
	a.Equal(w.Header().Get("Test"), "test")
}

func TestServer_buildPprof(t *testing.T) {
	a := assert.New(t)

	c := DefaultConfig()
	c.Pprof = "/pprof"
	s, err := New(c)
	a.NotError(err).NotNil(s)
	h := s.buildPprof(h1)

	// 命中 /pprof/cmdline
	r := httptest.NewRequest("GET", "http://caixw.io/pprof/cmdline", nil)
	w := httptest.NewRecorder()
	a.NotNil(r).NotNil(w)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusOK)

	// 命中 h1
	r = httptest.NewRequest("GET", "http://caixw.io/test", nil)
	w = httptest.NewRecorder()
	a.NotNil(r).NotNil(w)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)
}
