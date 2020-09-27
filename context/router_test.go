// SPDX-License-Identifier: MIT

package context

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

var f1 = func(ctx *Context) { ctx.Render(http.StatusOK, nil, nil) }

func TestPrefix(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.Handler(), nil)

	p := server.Prefix("/p")
	a.NotNil(p)

	path := "/path"
	a.NotError(p.Handle(path, f1, http.MethodGet, http.MethodDelete))
	srv.Get("/p" + path).Do().Status(http.StatusOK)
	srv.Delete("/p" + path).Do().Status(http.StatusOK)
	srv.Post("/p"+path, nil).Do().Status(http.StatusMethodNotAllowed)

	p.Post(path, f1)
	srv.Post("/p"+path, nil).Do().Status(http.StatusOK)

	p.Patch(path, f1)
	srv.Patch("/p"+path, nil).Do().Status(http.StatusOK)

	p.Options(path, "abc")
	srv.NewRequest(http.MethodOptions, "/p"+path).
		Do().
		Status(http.StatusOK).
		Header("allow", "abc")

	p.Remove(path, http.MethodDelete)
	srv.Delete("/p" + path).Do().Status(http.StatusMethodNotAllowed)

	// resource

	path = "/resources/{id}"
	res := p.Resource(path)
	res.Get(f1).Delete(f1)
	srv.Get("/p" + path).Do().Status(http.StatusOK)
	srv.Delete("/p" + path).Do().Status(http.StatusOK)

	res.Remove(http.MethodDelete)
	srv.Delete("/p" + path).Do().Status(http.StatusMethodNotAllowed)
	res.Remove(http.MethodGet)
	srv.Delete("/p" + path).Do().Status(http.StatusNotFound)
}

func TestResource(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)

	path := "/path"
	res := server.Resource(path)
	a.NotNil(res)

	srv := rest.NewServer(t, server.Handler(), nil)

	res.Get(f1)
	srv.Get(path).Do().Status(http.StatusOK)

	res.Delete(f1)
	srv.Delete(path).Do().Status(http.StatusOK)

	res.Patch(f1)
	srv.Patch(path, nil).Do().Status(http.StatusOK)

	res.Put(f1)
	srv.Put(path, nil).Do().Status(http.StatusOK)

	res.Post(f1)
	srv.Post(path, nil).Do().Status(http.StatusOK)

	res.Remove(http.MethodPost)
	srv.Post(path, nil).Do().Status(http.StatusMethodNotAllowed)

	res.Options("def")
	srv.NewRequest(http.MethodOptions, path).Do().Header("allow", "def")
}

func TestServer_Handle(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.Handler(), nil)

	path := "/path"
	a.NotError(server.Handle(path, f1, http.MethodGet, http.MethodDelete))
	srv.Get(path).Do().Status(http.StatusOK)
	srv.Delete(path).Do().Status(http.StatusOK)
	srv.Post(path, nil).Do().Status(http.StatusMethodNotAllowed)

	// 不指定请求方法，表示所有请求方法
	path = "/path1"
	a.NotError(server.Handle(path, f1))
	srv.Delete(path).Do().Status(http.StatusOK)
	srv.Patch(path, nil).Do().Status(http.StatusOK)

	path = "/path2"

	srv.Delete(path).Do().Status(http.StatusNotFound)

	server.Delete(path, f1)
	srv.Delete(path).Do().Status(http.StatusOK)

	server.Get(path, f1)
	srv.Get(path).Do().Status(http.StatusOK)

	server.Post(path, f1)
	srv.Post(path, nil).Do().Status(http.StatusOK)

	server.Patch(path, f1)
	srv.Patch(path, nil).Do().Status(http.StatusOK)

	server.Put(path, f1)
	srv.Put(path, nil).Do().Status(http.StatusOK)

	srv.NewRequest(http.MethodOptions, path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT")

	// 自定义 options
	server.Options(path, "abc")
	srv.NewRequest(http.MethodOptions, path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "abc")

	server.Remove(path, http.MethodOptions)
	srv.NewRequest(http.MethodOptions, path).
		Do().
		Status(http.StatusMethodNotAllowed)
}
