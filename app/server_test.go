// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestApp_run(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata", nil)
	a.NotError(err).NotNil(app)
	app.config = defaultConfig()
	app.config.Port = ":8083"
	app.config.Static = map[string]string{"/static": "./testdata/"}

	app.mux.GetFunc("/test", f1)

	go func() {
		err := app.run(nil)
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
	app, err := New("./testdata", nil)
	a.NotError(err).NotNil(app)
	app.config.Port = ":8083"
	app.config.HTTPS = true
	app.config.KeyFile = "./testdata/key.pem"
	app.config.CertFile = "./testdata/cert.pem"
	app.config.HTTPState = httpStateDisabled
	app.mux.GetFunc("/test", f1)

	go func() {
		err := app.run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 加载证书比较慢，需要等待 srv.run() 启动完毕，不同机器可能需要的时间会不同
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
	app, err := New("./testdata", nil)
	a.NotError(err).NotNil(app)
	app.config.Port = ":8083"
	app.config.HTTPS = true
	app.config.KeyFile = "./testdata/key.pem"
	app.config.CertFile = "./testdata/cert.pem"
	app.config.HTTPState = httpStateRedirect

	app.mux.GetFunc("/test", f1)

	go func() {
		err := app.run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 加载证书比较慢，需要等待 srv.run() 启动完毕，不同机器可能需要的时间会不同
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
	app, err := New("./testdata", nil)
	a.NotError(err).NotNil(app)
	app.config.Port = ":8083"
	app.config.HTTPS = true
	app.config.KeyFile = "./testdata/key.pem"
	app.config.CertFile = "./testdata/cert.pem"
	app.config.HTTPState = httpStateListen

	app.mux.GetFunc("/test", f1)

	go func() {
		err := app.run(nil)
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
