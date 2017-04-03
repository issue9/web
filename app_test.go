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

func newApp(a *assert.Assertion) *App {
	app, err := NewApp("./testdata")
	a.NotError(err).NotNil(app)

	return app
}

func TestApp_File(t *testing.T) {
	a := assert.New(t)

	app := newApp(a)
	a.Equal(app.File("test"), "testdata/test")
	a.Equal(app.File("test/file.jpg"), "testdata/test/file.jpg")
}

func TestNewApp(t *testing.T) {
	a := assert.New(t)

	app := newApp(a)
	a.Equal(app.configDir, "./testdata").
		NotNil(app.config).
		NotNil(app.server).
		NotNil(app.modules).
		Nil(app.handler).
		NotNil(app.content)
}

func TestApp(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)
	logs.SetWriter(logs.LevelError, os.Stderr, "[ERR]", log.LstdFlags)
	logs.SetWriter(logs.LevelInfo, os.Stderr, "[INFO]", log.LstdFlags)

	f1 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}
	shutdown := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
		if err := app.Shutdown(50 * time.Microsecond); err != nil {
			logs.Error("SHUTDOWN:", err)
		}
	}
	restart := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
		if err := app.Restart(50 * time.Microsecond); err != nil {
			logs.Error("RESTART:", err)
		}
	}

	// 只有将路由初始化放在 modules 中，才能在重启时，正确重新初始化路由。
	app.NewModule("init", func() error {
		app.Mux().GetFunc("/test", f1)
		app.Mux().GetFunc("/restart", restart)
		app.Mux().GetFunc("/shutdown", shutdown)
		return nil
	})

	go func() {
		// 不判断返回值，在被关闭或是重启时，会返回 http.ErrServerClosed 错误
		app.Run(nil)
	}()

	// 正常访问
	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp).Equal(resp.StatusCode, 1)

	// 重启之后，依然能访问
	resp, err = http.Get("http://localhost:8082/restart")
	a.NotError(err).NotNil(resp)
	time.Sleep(500 * time.Microsecond) // 待待 app.Restart 生效果
	resp, err = http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp).Equal(resp.StatusCode, 1)

	// 关闭
	resp, err = http.Get("http://localhost:8082/shutdown")
	a.NotError(err).NotNil(resp).Equal(resp.StatusCode, 1)
	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)
}
