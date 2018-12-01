// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert"
	yaml "gopkg.in/yaml.v2"

	"github.com/issue9/middleware/compress"
	"github.com/issue9/web/config"
	"github.com/issue9/web/mimetype"
	"github.com/issue9/web/mimetype/gob"
)

const timeout = 300 * time.Microsecond

var (
	f202 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}

	h202 = http.HandlerFunc(f202)
)

// 声明一个 App 实例
func newApp(a *assert.Assertion) *App {
	app, err := New(&Config{
		Dir: "./testdata",

		ConfigUnmarshals: map[string]config.UnmarshalFunc{
			".yaml": yaml.Unmarshal,
			".yml":  yaml.Unmarshal,
			".xml":  xml.Unmarshal,
			".json": json.Unmarshal,
		},

		Compresses: map[string]compress.WriterFunc{
			"gizp":    compress.NewGzip,
			"deflate": compress.NewDeflate,
		},

		MimetypeMarshals: map[string]mimetype.MarshalFunc{
			"application/json":       json.Marshal,
			"application/xml":        xml.Marshal,
			mimetype.DefaultMimetype: gob.Marshal,
		},

		MimetypeUnmarshals: map[string]mimetype.UnmarshalFunc{
			"application/json":       json.Unmarshal,
			"application/xml":        xml.Unmarshal,
			mimetype.DefaultMimetype: gob.Unmarshal,
		},
	})

	a.NotError(err).NotNil(app)

	return app
}

func TestApp_SetMiddleware(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)
	m := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Date", "1111")
			h.ServeHTTP(w, r)
		})
	}
	app.SetMiddlewares(m)
	a.Equal(len(app.middlewares), 1)

	app.Mux().GetFunc("/middleware", f202)
	go func() {
		err := app.Serve()
		a.ErrorType(err, http.ErrServerClosed, "错误，%v", err.Error())
	}()

	// 等待 Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Millisecond)

	// 正常访问
	resp, err := http.Get("http://localhost:8082/middleware")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.Header.Get("Date"), "1111")
	app.Close()
}

func TestApp_URL(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	a.Equal(app.URL("/abc"), "http://localhost:8082/abc")
	a.Equal(app.URL("abc/def"), "http://localhost:8082/abc/def")
	a.Equal(app.URL(""), "http://localhost:8082")
}

func TestApp_Handler(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	m1 := app.NewModule("m1", "m1 desc", "m2")
	a.NotNil(m1)
	m2 := app.NewModule("m2", "m2 desc")
	a.NotNil(m2)
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
	app := newApp(a)

	m1 := app.NewModule("m1", "m1 desc", "m2")
	a.NotNil(m1)
	m2 := app.NewModule("m2", "m2 desc")
	a.NotNil(m2)
	m1.GetFunc("/m1/test", f202)
	m2.GetFunc("/m2/test", f202)
	app.Mux().GetFunc("/mux/test", f202)

	go func() {
		err := app.Serve()
		a.ErrorType(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err.Error())
	}()
	time.Sleep(500 * time.Microsecond)

	a.NotError(app.Serve()) // 多次调用，之后的调用，都直接返回空值。

	resp, err := http.Get("http://localhost:8082/m1/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	resp, err = http.Get("http://localhost:8082/m2/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	resp, err = http.Get("http://localhost:8082/mux/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	// static 中定义的静态文件
	resp, err = http.Get("http://localhost:8082/client/file1.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	resp, err = http.Get("http://localhost:8082/client/dir/file2.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	// 不存在的文件
	resp, err = http.Get("http://localhost:8082/client/dir/not-exists.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusNotFound)

	app.Close()
}

func TestApp_Close(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

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
	app := newApp(a)
	app.webConfig.ShutdownTimeout = 0

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
	app := newApp(a)

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
