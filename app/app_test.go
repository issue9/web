// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert"

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
	app, err := New("./testdata")

	app.mt.AddMarshals(map[string]mimetype.MarshalFunc{
		"application/json":       json.Marshal,
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
	})

	app.mt.AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/json":       json.Unmarshal,
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
	})

	a.NotError(err).NotNil(app)

	// 以下内容由配置文件决定
	a.True(app.IsDebug()).
		True(len(app.Modules()) > 0) // 最起码有 web-core 模板

	return app
}

func TestApp_AddMiddleware(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)
	m := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Date", "1111")
			h.ServeHTTP(w, r)
		})
	}
	app.AddMiddlewares(m)
	a.True(app.webConfig.Debug).
		Equal(app.webConfig.Domain, "localhost")

	app.Mux().GetFunc("/middleware", f202)
	go func() {
		err := app.Serve()
		a.ErrorType(err, http.ErrServerClosed, "错误，%v", err.Error())
	}()

	// 等待 Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Millisecond)

	// 启动服务之后，再添加中间件，不会产生 panic
	a.NotPanic(func() { app.AddMiddlewares(m) })

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

func TestApp_Serve(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)
	app.AddErrorHandler(func(w http.ResponseWriter, status int) {
		w.WriteHeader(status)
		w.Write([]byte("error handler test"))
	}, http.StatusNotFound)

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

	a.ErrorType(app.Serve(), http.ErrServerClosed) // 多次调用

	resp, err := http.Get("http://localhost:8082/m1/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	resp, err = http.Get("http://localhost:8082/m2/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	resp, err = http.Get("http://localhost:8082/mux/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusAccepted)

	// not found
	// 返回 ErrorHandler 内容
	resp, err = http.Get("http://localhost:8082/mux/not-exists.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusNotFound)
	text, err := ioutil.ReadAll(resp.Body)
	a.NotError(err).NotNil(text)
	a.Equal(string(text), "error handler test")

	// static 中定义的静态文件
	resp, err = http.Get("http://localhost:8082/client/file1.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	resp, err = http.Get("http://localhost:8082/client/dir/file2.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	// 不存在的文件
	// 这是个 BUG，具体参考 https://github.com/issue9/web/issues/4 暂时忽略
	/*resp, err = http.Get("http://localhost:8082/client/dir/not-exists.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusNotFound)
	text, err = ioutil.ReadAll(resp.Body)
	a.NotError(err).NotNil(text)
	a.Equal(string(text), "error handler test")
	*/

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