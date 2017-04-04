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

func TestMain(m *testing.M) {
	Init("./testdata")

	os.Exit(m.Run())
}

func TestApp_File(t *testing.T) {
	a := assert.New(t)

	app, err := NewApp("./testdata")
	a.NotError(err).NotNil(app)

	a.Equal(app.File("test"), "testdata/test")
	a.Equal(app.File("test/file.jpg"), "testdata/test/file.jpg")

	// 全局函数
	a.Equal(File("test"), "testdata/test")
	a.Equal(File("test/file.jpg"), "testdata/test/file.jpg")
}

func TestNewApp(t *testing.T) {
	a := assert.New(t)

	app, err := NewApp("./testdata")
	a.NotError(err).NotNil(app)

	a.Equal(app.configDir, "./testdata").
		NotNil(app.config).
		NotNil(app.server).
		NotNil(app.modules).
		Nil(app.handler).
		NotNil(app.content)
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
	restart := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
		if err := Restart(50 * time.Microsecond); err != nil {
			logs.Error("RESTART:", err)
		}
	}

	// 只有将路由初始化放在 modules 中，才能在重启时，正确重新初始化路由。
	NewModule("init", func() error {
		Mux().GetFunc("/test", f1)
		Mux().GetFunc("/restart", restart)
		Mux().GetFunc("/shutdown", shutdown)
		return nil
	})

	go func() {
		// 不判断返回值，在被关闭或是重启时，会返回 http.ErrServerClosed 错误
		Run(nil)
	}()

	// 等待 Run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Millisecond)

	// 正常访问
	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp).Equal(resp.StatusCode, 1)

	// 重启之后，依然能访问
	resp, err = http.Get("http://localhost:8082/restart")
	a.NotError(err).NotNil(resp)
	time.Sleep(500 * time.Microsecond) // 待待 Restart 生效果
	resp, err = http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp).Equal(resp.StatusCode, 1)

	// 关闭
	resp, err = http.Get("http://localhost:8082/shutdown")
	a.NotError(err).NotNil(resp).Equal(resp.StatusCode, 1)
	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)
}
