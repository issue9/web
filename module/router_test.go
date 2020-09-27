// SPDX-License-Identifier: MIT

package module

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"

	"github.com/issue9/web/context"
)

var (
	f1 = func(ctx *context.Context) {
		ctx.Render(http.StatusOK, nil, nil)
	}
)

func TestResource(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	srv := rest.NewServer(t, server.ctxServer.Handler(), nil)
	m := server.NewModule("m1", "m1 desc")
	a.NotNil(m)
	p := m.Prefix("/p")
	a.NotNil(p)
	path := "/path"
	res := p.Resource(path)
	res.Delete(f1)
	res.Get(f1)
	res.Post(f1)
	res.Patch(f1)
	res.Put(f1)
	res.Options("abcdef")

	a.NotError(server.InitModules("", log.New(ioutil.Discard, "", 0)))

	srv.Delete("/p" + path).Do().Status(http.StatusOK)
	srv.Get("/p" + path).Do().Status(http.StatusOK)
	srv.Post("/p"+path, nil).Do().Status(http.StatusOK)
	srv.Patch("/p"+path, nil).Do().Status(http.StatusOK)
	srv.Put("/p"+path, nil).Do().Status(http.StatusOK)
	srv.NewRequest(http.MethodOptions, "/p"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "abcdef")
}

func TestPrefix(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	srv := rest.NewServer(t, server.ctxServer.Handler(), nil)
	m := server.NewModule("m1", "m1 desc")
	a.NotNil(m)
	p := m.Prefix("/p")
	a.NotNil(p)
	path := "/path"
	p.Delete(path, f1)
	p.Get(path, f1)
	p.Post(path, f1)
	p.Patch(path, f1)
	p.Put(path, f1)
	p.Options(path, "abcdef")

	a.NotError(server.InitModules("", log.New(ioutil.Discard, "", 0)))

	srv.Delete("/p" + path).Do().Status(http.StatusOK)
	srv.Get("/p" + path).Do().Status(http.StatusOK)
	srv.Post("/p"+path, nil).Do().Status(http.StatusOK)
	srv.Patch("/p"+path, nil).Do().Status(http.StatusOK)
	srv.Put("/p"+path, nil).Do().Status(http.StatusOK)
	srv.NewRequest(http.MethodOptions, "/p"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "abcdef")
}

func TestModule_Handle(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	srv := rest.NewServer(t, server.ctxServer.Handler(), nil)
	m := server.NewModule("m1", "m1 desc")
	a.NotNil(m)

	path := "/path"
	a.NotError(m.Handle(path, f1, http.MethodGet, http.MethodDelete))

	a.NotError(server.InitModules("", log.New(ioutil.Discard, "", 0)))

	srv.Get(path).Do().Status(http.StatusOK)
	srv.Delete(path).Do().Status(http.StatusOK)
	srv.Post(path, nil).Do().Status(http.StatusMethodNotAllowed)

	// 不指定请求方法，表示所有请求方法

	server = newServer(a)
	srv = rest.NewServer(t, server.ctxServer.Handler(), nil)
	m = server.NewModule("m1", "m1 desc")
	a.NotNil(m)
	path = "/path1"
	a.NotError(m.Handle(path, f1))

	a.NotError(server.InitModules("", log.New(ioutil.Discard, "", 0)))

	srv.Delete(path).Do().Status(http.StatusOK)
	srv.Patch(path, nil).Do().Status(http.StatusOK)

	// 各个请求方法

	server = newServer(a)
	srv = rest.NewServer(t, server.ctxServer.Handler(), nil)
	m = server.NewModule("m1", "m1 desc")
	a.NotNil(m)
	path = "/path2"
	m.Delete(path, f1)
	m.Get(path, f1)
	m.Post(path, f1)
	m.Patch(path, f1)
	m.Put(path, f1)

	a.NotError(server.InitModules("", log.New(ioutil.Discard, "", 0)))

	srv.Delete(path).Do().Status(http.StatusOK)
	srv.Get(path).Do().Status(http.StatusOK)
	srv.Post(path, nil).Do().Status(http.StatusOK)
	srv.Patch(path, nil).Do().Status(http.StatusOK)
	srv.Put(path, nil).Do().Status(http.StatusOK)
	srv.NewRequest(http.MethodOptions, path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT")
}
