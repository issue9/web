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

func TestNewServer(t *testing.T) {
	a := assert.New(t)

	srv, err := newServer(nil)
	a.Error(err).Nil(srv)

	srv, err = newServer(defaultConfig())
	a.NotError(err).NotNil(srv)
}

func TestServer_run(t *testing.T) {
	a := assert.New(t)

	conf := defaultConfig()
	conf.Port = ":8083"
	conf.Static = map[string]string{"/static": "./testdata/"}
	srv, err := newServer(conf)
	a.NotError(err).NotNil(srv)
	srv.mux.GetFunc("/test", f1)

	go func() {
		err := srv.run(nil)
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

	srv.shutdown(0)
}

func TestServer_shutdown(t *testing.T) {
	a := assert.New(t)

	conf := defaultConfig()
	conf.Port = ":8083"
	srv, err := newServer(conf)
	a.NotError(err).NotNil(srv)
	srv.mux.GetFunc("/test", f1)
	srv.mux.GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("closed"))
		srv.shutdown(0) // 手动调用接口关闭
	})

	go func() {
		err := srv.run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 等待 srv.run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(time.Second)

	resp, err := http.Get("http://localhost:8083/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	resp, err = http.Get("http://localhost:8083/close")
	a.Error(err).Nil(resp)

	resp, err = http.Get("http://localhost:8083/test")
	a.Error(err).Nil(resp)
}

func TestServer_shutdown_timeout(t *testing.T) {
	a := assert.New(t)

	conf := defaultConfig()
	conf.Port = ":8083"
	srv, err := newServer(conf)
	a.NotError(err).NotNil(srv)
	srv.mux.GetFunc("/test", f1)
	srv.mux.GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("closed"))
		srv.shutdown(30 * time.Microsecond)
	})

	go func() {
		err := srv.run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 等待 srv.run() 启动完毕，不同机器可能需要的时间会不同
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

func TestServer_httpStateDisabled(t *testing.T) {
	a := assert.New(t)

	conf := defaultConfig()
	conf.Port = ":8083"
	conf.HTTPS = true
	conf.KeyFile = "./testdata/key.pem"
	conf.CertFile = "./testdata/cert.pem"
	conf.HTTPState = httpStateDisabled
	srv, err := newServer(conf)
	a.NotError(err).NotNil(srv)
	srv.mux.GetFunc("/test", f1)

	go func() {
		err := srv.run(nil)
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

	srv.shutdown(0)
}

func TestServer_httpStateRedirect(t *testing.T) {
	a := assert.New(t)

	conf := defaultConfig()
	conf.Port = ":8083"
	conf.HTTPS = true
	conf.KeyFile = "./testdata/key.pem"
	conf.CertFile = "./testdata/cert.pem"
	conf.HTTPState = httpStateRedirect
	srv, err := newServer(conf)
	a.NotError(err).NotNil(srv)
	srv.mux.GetFunc("/test", f1)

	go func() {
		err := srv.run(nil)
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

	srv.shutdown(0)
}

func TestServer_httpStateListen(t *testing.T) {
	a := assert.New(t)

	conf := defaultConfig()
	conf.Port = ":8083"
	conf.HTTPS = true
	conf.KeyFile = "./testdata/key.pem"
	conf.CertFile = "./testdata/cert.pem"
	conf.HTTPState = httpStateListen
	srv, err := newServer(conf)
	a.NotError(err).NotNil(srv)
	srv.mux.GetFunc("/test", f1)

	go func() {
		err := srv.run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	// 加载证书比较慢，需要等待 srv.run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(time.Second)

	tlsconf := &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsconf}}
	resp, err := client.Get("https://localhost:8083/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	resp, err = http.Get("http://localhost:80/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	srv.shutdown(0)
}
