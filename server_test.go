// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

var f202 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	_, err := w.Write([]byte("1234567890"))
	if err != nil {
		println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// 声明一个 Server 实例
func newServer(a *assert.Assertion) *Web {
	conf, err := Classic("./testdata")
	a.NotError(err).NotNil(conf)

	// 以下内容由配置文件决定
	a.True(conf.Debug)

	return conf
}

func TestWeb_Run(t *testing.T) {
	a := assert.New(t)
	web := newServer(a)
	a.NotError(web.Init())
	exit := make(chan bool, 1)

	r := web.CTXServer().Router()
	r.Mux().GetFunc("/m1/test", f202)
	r.Mux().GetFunc("/m2/test", f202)
	r.Mux().GetFunc("/mux/test", f202)

	go func() {
		err := web.Serve()
		a.ErrorType(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err)
		exit <- true
	}()
	time.Sleep(500 * time.Microsecond) // 等待 go func() 完成

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/m1/test").
		Do().
		Status(http.StatusAccepted)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/m2/test").
		Do().
		Status(http.StatusAccepted)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/mux/test").
		Do().
		Status(http.StatusAccepted)

	// static 中定义的静态文件
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/admin/logs.xml").
		Do().
		Status(http.StatusOK)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/admin/logs.xml").
		Do().
		Status(http.StatusOK)

	a.NotError(web.Close())
	<-exit
}

func TestWeb_Close(t *testing.T) {
	a := assert.New(t)
	web := newServer(a)
	a.NotError(web.Init())
	exit := make(chan bool, 1)

	web.CTXServer().Router().Mux().GetFunc("/test", f202)
	web.CTXServer().Router().Mux().GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("closed"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		a.NotError(web.Close())
	})

	go func() {
		err := web.Serve()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
		exit <- true
	}()

	// 等待 srv.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Microsecond)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/test").
		Do().
		Status(http.StatusAccepted)

	// 连接被关闭，返回错误内容
	resp, err := http.Get("http://localhost:8082/close")
	a.Error(err).Nil(resp)

	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)

	<-exit
}

func TestWeb_Shutdown(t *testing.T) {
	a := assert.New(t)
	web := newServer(a)
	web.ShutdownTimeout = Duration(300 * time.Millisecond)
	a.NotError(web.Init())
	exit := make(chan bool, 1)

	web.CTXServer().Router().Mux().GetFunc("/test", f202)
	web.CTXServer().Router().Mux().GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte("shutdown with ctx"))
		a.NotError(err)
		web.Close()
	})

	go func() {
		err := web.Serve()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
		exit <- true
	}()

	// 等待 srv.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Microsecond)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/test").
		Do().
		Status(http.StatusAccepted)

	// 关闭指令可以正常执行
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/close").
		Do().
		Status(http.StatusCreated)

	// 未超时，但是拒绝新的链接
	resp, err := http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)

	// 已被关闭
	time.Sleep(30 * time.Microsecond)
	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)

	<-exit
}
