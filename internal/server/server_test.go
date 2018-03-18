// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/web/internal/config"
)

const timeout = 300 * time.Microsecond

var (
	h202 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

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

func TestListenClose(t *testing.T) {
	a := assert.New(t)
	conf := &config.Config{
		Port: 8080,
	}

	go func() {
		a.Equal(Listen(h202, conf), http.ErrServerClosed)
	}()
	time.Sleep(300 * time.Microsecond) // 等待启动完成

	// 正常访问
	resp, err := http.Get("http://localhost:8080/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	// 关闭
	a.NotError(Close())
	time.Sleep(300 * time.Microsecond)

	// 已经关闭，不能访问
	resp, err = http.Get("http://localhost:8080/test")
	a.Error(err).Nil(resp)
}

func testCloseWithTimeout(t *testing.T) {
	a := assert.New(t)
	conf := &config.Config{
		Port: 8080,
	}

	go func() {
		a.Equal(Listen(timeoutHandler, conf), http.ErrServerClosed)
	}()
	time.Sleep(300 * time.Microsecond) // 等待启动完成

	// 正常访问
	resp, err := http.Get("http://localhost:8080/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	// 等待以下两个协程完成
	wg := sync.WaitGroup{}
	wg.Add(2)
	defer wg.Wait()

	// 在执行过程中被关闭，不能访问
	go func() {
		resp, err = http.Get("http://localhost:8080/test")
		a.Error(err).Nil(resp)
		wg.Done()
	}()

	// 关闭
	go func() {
		a.NotError(Close())
		wg.Done()
	}()
}

func TestBuildHandler(t *testing.T) {
	a := assert.New(t)

	panicFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("err")
	})

	// 触发 panic
	conf = &config.Config{}
	h := buildHandler(panicFunc)
	request(a, h, "http://example.com/test", http.StatusInternalServerError)

	// 触发 panic，调试模式
	conf = &config.Config{
		Debug: true,
	}
	h = buildHandler(panicFunc)
	request(a, h, "http://example.com/test", http.StatusNotFound)
}

func TestBuildHosts_empty(t *testing.T) {
	a := assert.New(t)
	conf = &config.Config{}

	h := buildHosts(h202)
	request(a, h, "http://example.com/test", http.StatusAccepted)
}

func TestBuildHosts(t *testing.T) {
	a := assert.New(t)
	conf = &config.Config{
		AllowedDomains: []string{"caixw.io", "example.com"},
	}

	h := buildHosts(h202)

	// 带正确的域名访问
	request(a, h, "http://caixw.io/test", http.StatusAccepted)

	// 带不允许的域名访问
	request(a, h, "http://not.allowed/test", http.StatusNotFound)
}

func TestBuildHeader(t *testing.T) {
	a := assert.New(t)
	conf = &config.Config{
		Headers: map[string]string{"Test": "test"},
	}

	h := buildHeader(h202)

	r := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusAccepted)
	a.Equal(w.Header().Get("Test"), "test")
}

func TestBuildPprof(t *testing.T) {
	a := assert.New(t)
	conf = &config.Config{}

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
