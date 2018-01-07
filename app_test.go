// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/logs"
)

var initErr error

func TestMain(m *testing.M) {
	initErr = Init("./testdata", nil)

	os.Exit(m.Run())
}

// 检测在 TestMain() 中的功能是否存在错误。
func TestInit(t *testing.T) {
	a := assert.New(t)

	a.NotError(initErr).NotNil(defaultApp)
}

func TestBuildHandler(t *testing.T) {
	a := assert.New(t)
	app, err := NewApp("./testdata", func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Date", "1111")
			h.ServeHTTP(w, r)
		})
	})
	a.NotError(err).NotNil(app)

	f1 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}

	app.Router().GetFunc("/builder", f1)
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

func TestApp_File(t *testing.T) {
	a := assert.New(t)

	app, err := NewApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	a.Equal(app.File("test"), "testdata/test")
	a.Equal(app.File("test/file.jpg"), "testdata/test/file.jpg")

	// 全局函数
	a.Equal(File("test"), "testdata/test")
	a.Equal(File("test/file.jpg"), "testdata/test/file.jpg")
}

func TestURL(t *testing.T) {
	a := assert.New(t)

	a.Equal(URL("test"), "https://caixw.io/test")
	a.Equal(URL("/test/file.jpg"), "https://caixw.io/test/file.jpg")
}

func TestNewApp(t *testing.T) {
	a := assert.New(t)

	app, err := NewApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	a.Equal(app.configDir, "./testdata").
		NotNil(app.config).
		NotNil(app.server).
		NotNil(app.modules)
}

func TestApp(t *testing.T) {
	a := assert.New(t)
	logs.SetWriter(logs.LevelError, os.Stderr, "[ERR]", log.LstdFlags)
	logs.SetWriter(logs.LevelInfo, os.Stderr, "[INFO]", log.LstdFlags)

	f1 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}
	shutdown := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
		if err := Shutdown(50 * time.Microsecond); err != nil {
			logs.Error("SHUTDOWN:", err)
		}
	}

	Router().GetFunc("/out", f1)
	// 只有将路由初始化放在 modules 中，才能在重启时，正确重新初始化路由。
	Module("init", func() error {
		Router().GetFunc("/test", f1)
		Router().GetFunc("/shutdown", shutdown)
		return nil
	})

	go func() {
		// 不判断返回值，在被关闭或是重启时，会返回 http.ErrServerClosed 错误
		Run()
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
