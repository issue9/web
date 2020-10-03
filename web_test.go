// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

var f202 = func(ctx *Context) {
	ctx.Response.WriteHeader(http.StatusAccepted)
	_, err := ctx.Response.Write([]byte("1234567890"))
	if err != nil {
		println(err)
		ctx.Response.WriteHeader(http.StatusInternalServerError)
	}
}

// 声明一个 Server 实例
func newServer(a *assert.Assertion) *Web {
	web, err := Classic("./testdata")
	a.NotError(err).NotNil(web)

	return web
}

func TestWeb_Run(t *testing.T) {
	a := assert.New(t)
	exit := make(chan bool, 1)

	web := newServer(a)
	web.CTXServer().Get("/mux/test", f202)

	m1 := web.modServer.NewModule("m1", "m1 desc")
	m1.Get("/m1/test", f202)
	m1.NewTag("tag1")

	m2 := web.modServer.NewModule("m2", "m2 desc", "m1")
	m2.Get("/m2/test", func(ctx *Context) {
		w := GetWeb(ctx)
		a.NotNil(w)
		a.Equal(2, len(w.Modules()))
		a.Equal(2, len(w.Tags())).
			Equal(w.Tags()["m1"], []string{"tag1"}).
			Empty(w.Tags()["m2"])
		a.Equal(1, len(w.Services())) // 默认启动的 scheduled

		ctx.Response.WriteHeader(http.StatusAccepted)
		_, err := ctx.Response.Write([]byte("1234567890"))
		if err != nil {
			println(err)
			ctx.Response.WriteHeader(http.StatusInternalServerError)
		}
	})

	a.NotError(web.InitModules(""))

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
	exit := make(chan bool, 1)

	web.CTXServer().Get("/test", f202)
	web.CTXServer().Get("/close", func(ctx *Context) {
		_, err := ctx.Response.Write([]byte("closed"))
		if err != nil {
			ctx.Response.WriteHeader(http.StatusInternalServerError)
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
	web.shutdownTimeout = 300 * time.Millisecond
	exit := make(chan bool, 1)

	web.CTXServer().Get("/test", f202)
	web.CTXServer().Get("/close", func(ctx *Context) {
		ctx.Response.WriteHeader(http.StatusCreated)
		_, err := ctx.Response.Write([]byte("shutdown with ctx"))
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
