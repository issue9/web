// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/logs"
)

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(1)
}

var h1 = http.HandlerFunc(f1)

func TestBuildHandler(t *testing.T) {
	a := assert.New(t)
	conf := defaultConfig()
	conf.Build = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Date", "1111")
			h.ServeHTTP(w, r)
		})
	}
	app, err := New(conf)
	a.NotError(err).NotNil(app)

	f1 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}

	app.router.GetFunc("/builder", f1)
	go func() {
		// 不判断返回值，在被关闭或是重启时，会返回 http.ErrServerClosed 错误
		app.Run()
	}()

	// 等待 Run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Millisecond)

	// 正常访问
	resp, err := http.Get("http://localhost:8082/builder")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.Header.Get("Date"), "1111")
	app.Shutdown(0)
}

func TestApp_initFromConfig(t *testing.T) {
	a := assert.New(t)

	app := &App{}
	conf := defaultConfig()
	conf.HTTPS = true
	conf.Port = httpsPort
	conf.Domain = "example.com"
	conf.Root = "/path"
	app.initFromConfig(conf)
	a.Equal(app.url, "https://example.com/path")

	app = &App{}
	conf.HTTPS = false
	app.initFromConfig(conf)
	a.Equal(app.url, "http://example.com:443/path")

	app = &App{}
	conf.HTTPS = false
	conf.Port = 80
	conf.Root = ""
	app.initFromConfig(conf)
	a.Equal(app.url, "http://example.com")
}

func TestApp_URL(t *testing.T) {
	a := assert.New(t)

	app := &App{}
	conf := defaultConfig()
	conf.HTTPS = true
	conf.Port = 443
	conf.Domain = "example.com"
	app.initFromConfig(conf)

	a.Equal(app.URL("test"), "https://example.com/test")
	a.Equal(app.URL("/test/file.jpg"), "https://example.com/test/file.jpg")
}

func TestApp(t *testing.T) {
	a := assert.New(t)
	logs.SetWriter(logs.LevelError, os.Stderr, "[ERR]", log.LstdFlags)
	logs.SetWriter(logs.LevelInfo, os.Stderr, "[INFO]", log.LstdFlags)

	conf := defaultConfig()
	app, err := New(conf)
	a.NotError(err).NotNil(app)

	f1 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}
	shutdown := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
		if err := app.Shutdown(5 * time.Microsecond); err != nil {
			logs.Error("SHUTDOWN:", err)
		}
	}

	app.router.GetFunc("/out", f1)
	app.AddModule(&Module{
		Name:        "init",
		Description: "init 测试用",
		Routes: []*Route{
			{
				Methods: []string{http.MethodGet},
				Path:    "/test",
				Handler: http.HandlerFunc(f1),
			},
			{
				Methods: []string{http.MethodGet},
				Path:    "/shutdown",
				Handler: http.HandlerFunc(shutdown),
			},
		}})

	go func() {
		// 不判断返回值，在被关闭或是重启时，会返回 http.ErrServerClosed 错误
		app.Run()
	}()

	// 等待 Run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Millisecond)

	// 正常访问
	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp).Equal(resp.StatusCode, 1)
	resp, err = http.Get("http://localhost:8082/out")
	a.NotError(err).NotNil(resp).Equal(resp.StatusCode, 1)

	// 关闭
	resp, err = http.Get("http://localhost:8082/shutdown")
	a.NotError(err).NotNil(resp).Equal(resp.StatusCode, 1)
	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)
}

func TestApp_Shutdown(t *testing.T) {
	a := assert.New(t)
	config := defaultConfig()
	config.Port = 8083
	app := &App{}
	app.initFromConfig(config)

	app.mux.GetFunc("/test", f1)
	app.mux.GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("closed"))
		app.Shutdown(0) // 手动调用接口关闭
	})

	go func() {
		err := app.Run()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 等待 app.run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(50 * time.Microsecond)

	resp, err := http.Get("http://localhost:8083/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	resp, err = http.Get("http://localhost:8083/close")
	a.Error(err).Nil(resp)

	resp, err = http.Get("http://localhost:8083/test")
	a.Error(err).Nil(resp)
}

func TestApp_Shutdown_timeout(t *testing.T) {
	a := assert.New(t)
	config := defaultConfig()
	config.Port = 8083
	app := &App{}
	app.initFromConfig(config)

	app.mux.GetFunc("/test", f1)
	app.mux.GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("closed"))
		app.Shutdown(30 * time.Microsecond)
	})

	go func() {
		err := app.Run()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 等待 app.run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(50 * time.Microsecond)

	resp, err := http.Get("http://localhost:8083/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	// 关闭指令可以正常执行
	resp, err = http.Get("http://localhost:8083/close")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusCreated)

	// 拒绝访问
	resp, err = http.Get("http://localhost:8083/test")
	a.Error(err).Nil(resp)

	// 已被关闭
	time.Sleep(30 * time.Microsecond)
	resp, err = http.Get("http://localhost:8083/test")
	a.Error(err).Nil(resp)
}

func TestApp_Run(t *testing.T) {
	a := assert.New(t)
	config := defaultConfig()
	config.Port = 8083
	config.Static = map[string]string{"/static": "./testdata/"}
	app := &App{}
	app.initFromConfig(config)

	app.mux.GetFunc("/test", f1)

	go func() {
		err := app.Run()
		a.ErrorType(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err.Error())
	}()

	time.Sleep(50 * time.Microsecond)
	resp, err := http.Get("http://localhost:8083/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	resp, err = http.Get("http://localhost:8083/static/file1.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	resp, err = http.Get("http://localhost:8083/static/dir/file2.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	app.Shutdown(0)
}

func TestApp_NewContext(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()
	conf := defaultConfig()
	app, err := New(conf)
	a.NotError(err).NotNil(app)

	// 缺少 Accept 报头
	app.config.Strict = true
	ctx := app.NewContext(w, r)
	a.Nil(ctx)
	a.Equal(w.Code, http.StatusNotAcceptable)

	// 不检测 Accept 报头
	app.config.Strict = false
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	w = httptest.NewRecorder()
	ctx = app.NewContext(w, r)
	a.NotNil(ctx)
	a.Equal(w.Code, http.StatusOK)
}

func TestApp_buildHandler(t *testing.T) {
	a := assert.New(t)
	conf := defaultConfig()
	app, err := New(conf)
	a.NotError(err).NotNil(app)

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
	conf := defaultConfig()
	app, err := New(conf)
	a.NotError(err).NotNil(app)

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
	config.AllowedDomains = []string{"caixw.io", "example.com"} // 指定域名
	app := &App{}
	app.initFromConfig(config)

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

func TestApp_buildHeader(t *testing.T) {
	a := assert.New(t)
	config := defaultConfig()
	config.Headers = map[string]string{"Test": "test"}
	app := &App{}
	app.initFromConfig(config)

	h := app.buildHeader(h1)

	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)
	a.Equal(w.Header().Get("Test"), "test")
}

func TestApp_buildPprof(t *testing.T) {
	a := assert.New(t)

	conf := defaultConfig()
	app, err := New(conf)
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
