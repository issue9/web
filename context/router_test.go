// SPDX-License-Identifier: MIT

package context

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

var f1 = func(ctx *Context) { ctx.Render(http.StatusOK, nil, nil) }

func TestBuilder_Prefix(t *testing.T) {
	a := assert.New(t)
	builder := newServer(a)
	srv := rest.NewServer(t, builder.Handler(), nil)

	p := builder.Prefix("/p")
	a.NotNil(p)

	path := "/path"
	a.NotError(p.Handle(path, f1, http.MethodGet, http.MethodDelete))
	srv.NewRequest(http.MethodGet, "/p"+path).
		Do().
		Status(http.StatusOK)
	srv.NewRequest(http.MethodDelete, "/p"+path).
		Do().
		Status(http.StatusOK)
	srv.NewRequest(http.MethodPost, "/p"+path).
		Do().
		Status(http.StatusMethodNotAllowed)

	p.Post(path, f1)
	srv.NewRequest(http.MethodPost, "/p"+path).
		Do().
		Status(http.StatusOK)

	p.Patch(path, f1)
	srv.NewRequest(http.MethodPatch, "/p"+path).
		Do().
		Status(http.StatusOK)

	p.Options(path, f1)
	srv.NewRequest(http.MethodOptions, "/p"+path).
		Do().
		Status(http.StatusOK)
}

func TestBuilder_Handle(t *testing.T) {
	a := assert.New(t)
	builder := newServer(a)
	srv := rest.NewServer(t, builder.Handler(), nil)

	path := "/path"
	a.NotError(builder.Handle(path, f1, http.MethodGet, http.MethodDelete))
	srv.NewRequest(http.MethodGet, path).
		Do().
		Status(http.StatusOK)
	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusOK)
	srv.NewRequest(http.MethodPost, path).
		Do().
		Status(http.StatusMethodNotAllowed)

	// 不指定请求方法，表示所有请求方法
	path = "/path1"
	a.NotError(builder.Handle(path, f1))
	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusOK)
	srv.NewRequest(http.MethodPatch, path).
		Do().
		Status(http.StatusOK)
}

func TestBuilder_Handles(t *testing.T) {
	a := assert.New(t)
	builder := newServer(a)
	srv := rest.NewServer(t, builder.Handler(), nil)

	path := "/path"

	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusNotFound)

	builder.Get(path, f1)
	srv.NewRequest(http.MethodGet, path).
		Do().
		Status(http.StatusOK)

	builder.Post(path, f1)
	srv.NewRequest(http.MethodPost, path).
		Do().
		Status(http.StatusOK)

	builder.Patch(path, f1)
	srv.NewRequest(http.MethodPatch, path).
		Do().
		Status(http.StatusOK)

	builder.Put(path, f1)
	srv.NewRequest(http.MethodPut, path).
		Do().
		Status(http.StatusOK)

	builder.Delete(path, f1)
	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusOK)
}
