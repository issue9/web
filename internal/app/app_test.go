// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert"
)

var (
	f202 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}

	h202 = http.HandlerFunc(f202)
)

func TestMiddleware(t *testing.T) {
	a := assert.New(t)
	m := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Date", "1111")
			h.ServeHTTP(w, r)
		})
	}
	app, err := New("./testdata", m)
	a.NotError(err).NotNil(app)

	app.router.GetFunc("/middleware", f202)
	go func() {
		a.Equal(app.Run(), http.ErrServerClosed)
	}()

	// 等待 Run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Millisecond)

	// 正常访问
	resp, err := http.Get("http://localhost:8082/middleware")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.Header.Get("Date"), "1111")
	app.Close()
}

func TestApp_URL(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata", nil)
	a.NotError(err).NotNil(app)

	a.Equal(app.URL("/abc"), "http://localhost:8082/abc")
	a.Equal(app.URL("abc/def"), "http://localhost:8082/abc/def")
	a.Equal(app.URL(""), "http://localhost:8082")
}

func TestApp_NewContext(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	app, err := New("./testdata", nil)
	a.NotError(err).NotNil(app)

	// 少报头 accept
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "not")
	a.Panic(func() {
		app.NewContext(w, r)
	})

	// 少 accept-charset
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "*/*")
	r.Header.Set("Accept-Charset", "unknown")
	a.Panic(func() {
		app.NewContext(w, r)
	})

	// 正常
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "*/*")
	r.Header.Set("Accept-Charset", "*")
	ctx := app.NewContext(w, r)
	a.NotNil(ctx)
}
