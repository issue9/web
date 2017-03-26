// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestNew(t *testing.T) {
	a := assert.New(t)

	srv, err := New(nil)
	a.Error(err).Nil(srv)

	srv, err = New(DefaultConfig())
	a.NotError(err).NotNil(srv)
}

func TestServer_Run(t *testing.T) {
	a := assert.New(t)

	conf := DefaultConfig()
	conf.Port = ":8082"
	conf.Static = map[string]string{"/static": "./testdata/"}
	srv, err := New(conf)
	a.NotError(err).NotNil(srv)
	srv.Mux().GetFunc("/test", f1)

	go func() {
		err := srv.Run(nil)
		a.ErrorType(err, http.ErrServerClosed)
	}()

	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	resp, err = http.Get("http://localhost:8082/static/file1.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	resp, err = http.Get("http://localhost:8082/static/dir/file2.txt")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	srv.Shutdown(0)
}

func TestServer_Shutdown(t *testing.T) {
	a := assert.New(t)

	conf := DefaultConfig()
	conf.Port = ":8082"
	srv, err := New(conf)
	a.NotError(err).NotNil(srv)
	srv.Mux().GetFunc("/test", f1)
	srv.Mux().GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("closed"))
		srv.Shutdown(0) // 手动调用接口关闭
	})

	go func() {
		err := srv.Run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed)
	}()

	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	resp, err = http.Get("http://localhost:8082/close")
	a.Error(err).Nil(resp)

	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)
}

func TestServer_Shutdown_timeout(t *testing.T) {
	a := assert.New(t)

	conf := DefaultConfig()
	conf.Port = ":8082"
	srv, err := New(conf)
	a.NotError(err).NotNil(srv)
	srv.Mux().GetFunc("/test", f1)
	srv.Mux().GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("closed"))
		srv.Shutdown(30 * time.Microsecond)
	})

	go func() {
		err := srv.Run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed)
	}()

	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	// 关闭指令可以正常执行
	resp, err = http.Get("http://localhost:8082/close")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusCreated)

	// 拒绝访问
	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)

	// 已被关闭
	time.Sleep(30 * time.Microsecond)
	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)
}

func TestServer_httpStateDisabled(t *testing.T) {
	a := assert.New(t)

	conf := DefaultConfig()
	conf.Port = ":8083"
	conf.HTTPS = true
	conf.KeyFile = "./testdata/key.pem"
	conf.CertFile = "./testdata/cert.pem"
	conf.HTTPState = httpStateDisabled
	srv, err := New(conf)
	a.NotError(err).NotNil(srv)
	srv.Mux().GetFunc("/test", f1)

	go func() {
		err := srv.Run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed)
	}()

	// 加载证书比较慢，需要等待 srv.Run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Millisecond)

	tlsconf := &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsconf}}
	resp, err := client.Get("https://localhost:8083/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	resp, err = http.Get("http://localhost:8083/test")
	a.Error(err).Nil(resp)

	srv.Shutdown(0)
}

func TestServer_httpStateRedirect(t *testing.T) {
	a := assert.New(t)

	conf := DefaultConfig()
	conf.Port = ":8083"
	conf.HTTPS = true
	conf.KeyFile = "./testdata/key.pem"
	conf.CertFile = "./testdata/cert.pem"
	conf.HTTPState = httpStateRedirect
	srv, err := New(conf)
	a.NotError(err).NotNil(srv)
	srv.Mux().GetFunc("/test", f1)

	go func() {
		err := srv.Run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed)
	}()

	// 加载证书比较慢，需要等待 srv.Run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Millisecond)

	tlsconf := &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsconf}}
	resp, err := client.Get("https://localhost:8083/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	resp, err = client.Get("http://localhost:80/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	srv.Shutdown(0)
}

func TestServer_httpStateListen(t *testing.T) {
	a := assert.New(t)

	conf := DefaultConfig()
	conf.Port = ":8083"
	conf.HTTPS = true
	conf.KeyFile = "./testdata/key.pem"
	conf.CertFile = "./testdata/cert.pem"
	conf.HTTPState = httpStateListen
	srv, err := New(conf)
	a.NotError(err).NotNil(srv)
	srv.Mux().GetFunc("/test", f1)

	go func() {
		err := srv.Run(nil)
		a.Error(err).ErrorType(err, http.ErrServerClosed)
	}()

	// 加载证书比较慢，需要等待 srv.Run() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Millisecond)

	tlsconf := &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsconf}}
	resp, err := client.Get("https://localhost:8083/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	resp, err = http.Get("http://localhost:80/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	srv.Shutdown(0)
}
