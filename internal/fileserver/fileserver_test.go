// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fileserver

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/middleware"

	"github.com/issue9/web/internal/exit"
)

// New 错误返回
func TestFileServer_faild(t *testing.T) {
	a := assert.New(t)

	h := New("./testdata")
	h = middleware.Handler(h, func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				msg, ok := recover().(exit.HTTPStatus)
				a.True(ok).Equal(msg, http.StatusNotFound)
			}()

			h.ServeHTTP(w, r)
		})
	})

	srv := httptest.NewServer(h)
	resp, err := http.Get(srv.URL + "/not-exists")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)
}

// New 正常返回的
func TestFileServer_ok(t *testing.T) {
	a := assert.New(t)

	h := New("./testdata")
	h = middleware.Handler(h, func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				msg := recover()
				a.Nil(msg)
			}()

			h.ServeHTTP(w, r)
		})
	})

	srv := httptest.NewServer(h)
	resp, err := http.Get(srv.URL + "/file")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)
}

// ServeFile 错误返回
func TestServeFile_faild(t *testing.T) {
	a := assert.New(t)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			msg, ok := recover().(exit.HTTPStatus)
			a.True(ok).Equal(msg, http.StatusNotFound)
		}()

		ServeFile(w, r, "./test/not-exists")
	})

	srv := httptest.NewServer(h)
	resp, err := http.Get(srv.URL)
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)
}

// ServeFile 正常返回的
func TestServeFile_ok(t *testing.T) {
	a := assert.New(t)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			msg := recover()
			a.Nil(msg)
		}()

		ServeFile(w, r, "./testdata/file")
	})

	srv := httptest.NewServer(h)
	resp, err := http.Get(srv.URL)
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)
}

// ServeContent 错误返回
func TestServeContent_ok(t *testing.T) {
	a := assert.New(t)
	buf, err := ioutil.ReadFile("./testdata/file")
	a.NotError(err).NotNil(buf)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			msg := recover()
			a.Nil(msg)
		}()

		ServeContent(w, r, "test", time.Now(), bytes.NewReader(buf))
	})

	srv := httptest.NewServer(h)
	resp, err := http.Get(srv.URL)
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)
}
