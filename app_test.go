// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert"
)

var f202 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}

var h202 = http.HandlerFunc(f202)

func TestMiddleware(t *testing.T) {
	a := assert.New(t)
	m := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Date", "1111")
			h.ServeHTTP(w, r)
		})
	}
	app, err := newApp("./testdata", m)
	a.NotError(err).NotNil(app)

	app.router.GetFunc("/middleware", f202)
	go func() {
		// 不判断返回值，在被关闭或是重启时，会返回 http.ErrServerClosed 错误
		app.Run()
	}()

	// 等待 Run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Millisecond)

	// 正常访问
	resp, err := http.Get("http://localhost:8082/middleware")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.Header.Get("Date"), "1111")
	Close()
}

func TestApp_Close(t *testing.T) {
	a := assert.New(t)
	app, err := newApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	app.mux.GetFunc("/test", f202)
	app.mux.GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("closed"))
		Close()
	})

	go func() {
		err := app.Run()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 等待 app.run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Microsecond)

	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	resp, err = http.Get("http://localhost:8082/close")
	a.Error(err).Nil(resp)

	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)
}

func TestApp_Shutdown_timeout(t *testing.T) {
	a := assert.New(t)
	app, err := newApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	app.mux.GetFunc("/test", f202)
	app.mux.GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("closed"))
		Shutdown()
	})

	go func() {
		err := app.Run()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 等待 app.run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Microsecond)

	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	// 关闭指令可以正常执行
	resp, err = http.Get("http://localhost:8082/close")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusCreated)

	// 拒绝访问
	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)

	// 已被关闭
	time.Sleep(30 * time.Microsecond)
	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)
}

func TestApp_Run(t *testing.T) {
	a := assert.New(t)
	app, err := newApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	app.mux.GetFunc("/test", f202)

	go func() {
		err := app.Run()
		a.ErrorType(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err.Error())
	}()

	time.Sleep(500 * time.Microsecond)
	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	resp, err = http.Get("http://localhost:8082/client/file1.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	resp, err = http.Get("http://localhost:8082/client/dir/file2.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	Close()
}

func TestApp_NewContext(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	app, err := newApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	// 少报头 accept
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "not")
	ctx := app.NewContext(w, r)
	a.Nil(ctx)

	// 少 accept-charset
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "*/*")
	r.Header.Set("Accept-Charset", "unknown")
	ctx = app.NewContext(w, r)
	a.Nil(ctx)

	// 正常
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "*/*")
	r.Header.Set("Accept-Charset", "*")
	ctx = app.NewContext(w, r)
	a.NotNil(ctx)
}
