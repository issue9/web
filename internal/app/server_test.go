// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/web/context"
	"github.com/issue9/web/internal/config"
)

const timeout = 300 * time.Microsecond

var (
	timeoutHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(timeout)
		w.WriteHeader(http.StatusAccepted)

	})
)

// 验证请求地址是否返回正确的状态码
func request(a *assert.Assertion, h http.Handler, url string, code int) {
	r := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, code)
}

func TestApp_Handler(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	m1, err := app.NewModule("m1", "m1 desc", "m2")
	a.NotError(err).NotNil(m1)
	m2, err := app.NewModule("m2", "m2 desc")
	a.NotError(err).NotNil(m2)
	m1.GetFunc("/m1/test", f202)
	m2.GetFunc("/m2/test", f202)

	h1, err := app.Handler()
	a.NotError(err).NotNil(h1)

	h2, err := app.Handler()
	a.NotError(err).NotNil(h2)

	a.Equal(h1, h2)
}

func TestApp_Serve(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	m1, err := app.NewModule("m1", "m1 desc", "m2")
	a.NotError(err).NotNil(m1)
	m2, err := app.NewModule("m2", "m2 desc")
	a.NotError(err).NotNil(m2)
	m1.GetFunc("/m1/test", f202)
	m2.GetFunc("/m2/test", f202)

	go func() {
		err := app.Serve()
		a.ErrorType(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err.Error())
	}()
	time.Sleep(500 * time.Microsecond)

	a.NotError(app.Serve()) // 多次调用，之后的调用，都直接返回空值。

	resp, err := http.Get("http://localhost:8082/m1/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	resp, err = http.Get("http://localhost:8082/client/file1.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	resp, err = http.Get("http://localhost:8082/client/dir/file2.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	app.Close()
}

func TestApp_Close(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	app.mux.GetFunc("/test", f202)
	app.mux.GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("closed"))
		app.Close()
	})

	// 未调用 Serve 时，调用 Close，应该不会有任何变化
	a.NotError(app.Close())

	go func() {
		err := app.Serve()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 等待 app.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Microsecond)

	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	resp, err = http.Get("http://localhost:8082/close")
	a.Error(err).Nil(resp)

	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)
}

func TestApp_shutdown(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	app.config.ShutdownTimeout = 0
	a.NotError(err).NotNil(app)

	app.mux.GetFunc("/test", f202)
	app.mux.GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("shutdown"))
		app.Shutdown()
	})

	// 未调用 Serve 时，调用 Close，应该不会有任何变化
	a.NotError(app.Shutdown())

	go func() {
		err := app.Serve()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 等待 app.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Microsecond)

	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	// 调用关闭操作
	resp, err = http.Get("http://localhost:8082/close")
	a.Error(err).Nil(resp)

	// 立即关闭
	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)
}

func TestApp_Shutdown_timeout(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	app.mux.GetFunc("/test", f202)
	app.mux.GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("shutdown with timeout"))
		app.Shutdown()
	})

	// 未调用 Serve 时，调用 Shutdown，应该不会有任何变化
	a.NotError(app.Shutdown())

	go func() {
		err := app.Serve()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 等待 app.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Microsecond)

	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	// 关闭指令可以正常执行
	resp, err = http.Get("http://localhost:8082/close")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusCreated)

	// 未超时，但是拒绝新的链接
	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)

	// 已被关闭
	time.Sleep(30 * time.Microsecond)
	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)
}

func TestBuildHandler(t *testing.T) {
	a := assert.New(t)

	panicFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("err")
	})

	panicHTTPFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context.Exit(http.StatusNotAcceptable)
	})

	// 触发 panic
	h := buildHandler(&config.Config{}, panicFunc)
	request(a, h, "http://example.com/test", http.StatusInternalServerError)

	// 触发 panic，调试模式
	conf := &config.Config{
		Debug: true,
	}
	h = buildHandler(conf, panicFunc)
	request(a, h, "http://example.com/test", http.StatusInternalServerError)

	// 触发 panic, errors.HTTP
	h = buildHandler(&config.Config{}, panicHTTPFunc)
	request(a, h, "http://example.com/test", http.StatusNotAcceptable)
}

func TestBuildHosts_empty(t *testing.T) {
	a := assert.New(t)

	h := buildHosts(&config.Config{}, h202)
	request(a, h, "http://example.com/test", http.StatusAccepted)
}

func TestBuildHosts(t *testing.T) {
	a := assert.New(t)
	conf := &config.Config{
		AllowedDomains: []string{"caixw.io", "example.com"},
	}

	h := buildHosts(conf, h202)

	// 带正确的域名访问
	request(a, h, "http://caixw.io/test", http.StatusAccepted)

	// 带不允许的域名访问
	request(a, h, "http://not.allowed/test", http.StatusNotFound)
}

func TestBuildHeader(t *testing.T) {
	a := assert.New(t)
	conf := &config.Config{
		Headers: map[string]string{"Test": "test"},
	}

	h := buildHeader(conf, h202)

	r := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusAccepted)
	a.Equal(w.Header().Get("Test"), "test")
}

func TestBuildPprof(t *testing.T) {
	a := assert.New(t)
	h := buildPprof(h202)

	// 命中 /debug/pprof/cmdline
	request(a, h, "http://example.com/debug/pprof/", http.StatusOK)
	request(a, h, "http://example.com/debug/pprof/cmdline", http.StatusOK)
	request(a, h, "http://example.com/debug/pprof/trace", http.StatusOK)
	request(a, h, "http://example.com/debug/pprof/symbol", http.StatusOK)
	//request(a, h, "http://example.com/debug/pprof/profile", http.StatusOK)

	// 命中 h202
	request(a, h, "http://example.com/debug", http.StatusAccepted)
}
