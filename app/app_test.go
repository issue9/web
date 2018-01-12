// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/logs"
)

func TestBuildHandler(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	f1 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}

	app.Router().GetFunc("/builder", f1)
	go func() {
		// 不判断返回值，在被关闭或是重启时，会返回 http.ErrServerClosed 错误
		app.Run(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Date", "1111")
				h.ServeHTTP(w, r)
			})
		})
	}()

	// 等待 Run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Millisecond)

	// 正常访问
	resp, err := http.Get("http://localhost:8082/builder")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.Header.Get("Date"), "1111")
	app.Shutdown(0)
}

func TestNew(t *testing.T) {
	a := assert.New(t)

	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	f, err := filepath.Abs("./testdata")
	a.NotError(err)
	a.Equal(app.configDir, f).
		NotNil(app.config).
		NotNil(app.modules)

	app, err = New("./not-exists")
	a.Error(err).Nil(app)
}

func TestApp_File(t *testing.T) {
	a := assert.New(t)

	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	f, err := filepath.Abs("./testdata/test")
	a.NotError(err)
	a.Equal(app.File("test"), f)

	f, err = filepath.Abs("./testdata/test/file.jpg")
	a.NotError(err)
	a.Equal(app.File("test/file.jpg"), f)
}

func TestApp_URL(t *testing.T) {
	a := assert.New(t)

	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	a.Equal(app.URL("test"), "https://caixw.io/test")
	a.Equal(app.URL("/test/file.jpg"), "https://caixw.io/test/file.jpg")
}

func TestApp(t *testing.T) {
	a := assert.New(t)
	logs.SetWriter(logs.LevelError, os.Stderr, "[ERR]", log.LstdFlags)
	logs.SetWriter(logs.LevelInfo, os.Stderr, "[INFO]", log.LstdFlags)

	app, err := New("./testdata")
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

	app.Router().GetFunc("/out", f1)
	app.Module("init", func() error {
		app.Router().GetFunc("/test", f1)
		app.Router().GetFunc("/shutdown", shutdown)
		return nil
	})

	go func() {
		// 不判断返回值，在被关闭或是重启时，会返回 http.ErrServerClosed 错误
		app.Run(nil)
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
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)
	app.config = defaultConfig()
	app.config.Port = 8083

	app.mux.GetFunc("/test", f1)
	app.mux.GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("closed"))
		app.Shutdown(0) // 手动调用接口关闭
	})

	go func() {
		err := app.Run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 等待 app.run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(time.Second)

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
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)
	app.config = defaultConfig()
	app.config.Port = 8083

	app.mux.GetFunc("/test", f1)
	app.mux.GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("closed"))
		app.Shutdown(30 * time.Microsecond)
	})

	go func() {
		err := app.Run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 等待 app.run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(time.Second)

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
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)
	app.config = defaultConfig()
	app.config.Port = 8083
	app.config.Static = map[string]string{"/static": "./testdata/"}

	app.mux.GetFunc("/test", f1)

	go func() {
		err := app.Run(nil)
		a.ErrorType(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err.Error())
	}()

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

func TestApp_httpStateDisabled(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)
	app.config.Port = 8083
	app.config.HTTPS = true
	app.config.KeyFile = "./testdata/key.pem"
	app.config.CertFile = "./testdata/cert.pem"
	app.config.HTTPState = httpStateDisabled
	app.mux.GetFunc("/test", f1)

	go func() {
		err := app.Run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 加载证书比较慢，需要等待 app.run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(time.Second)

	tlsconf := &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsconf}}
	resp, err := client.Get("https://localhost:8083/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	resp, err = http.Get("http://localhost:8083/test")
	a.Error(err).Nil(resp)

	app.Shutdown(0)
}

func TestApp_httpStateRedirect(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)
	app.config.Port = 8083
	app.config.HTTPS = true
	app.config.KeyFile = "./testdata/key.pem"
	app.config.CertFile = "./testdata/cert.pem"
	app.config.HTTPState = httpStateRedirect

	app.mux.GetFunc("/test", f1)

	go func() {
		err := app.Run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 加载证书比较慢，需要等待 app.run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(time.Second)

	tlsconf := &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsconf}}
	resp, err := client.Get("https://localhost:8083/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	resp, err = client.Get("http://localhost:80/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	app.Shutdown(0)
}

func TestApp_httpStateListen(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)
	app.config.Port = 8083
	app.config.HTTPS = true
	app.config.KeyFile = "./testdata/key.pem"
	app.config.CertFile = "./testdata/cert.pem"
	app.config.HTTPState = httpStateListen

	app.mux.GetFunc("/test", f1)

	go func() {
		err := app.Run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 加载证书比较慢，需要等待 app.run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(time.Second)

	tlsconf := &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsconf}}
	resp, err := client.Get("https://localhost:8083/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	resp, err = http.Get("http://localhost:80/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	app.Shutdown(0)
}
