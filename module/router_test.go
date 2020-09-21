// SPDX-License-Identifier: MIT

package module

import (
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

func TestModule_Prefix(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)

	srv := rest.NewServer(t, server.ctxServer.Handler(), nil)

	m := server.newModule("m1", "m1 desc")
	a.NotNil(m)
	p := m.Prefix("/p")
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
}

func TestModule_Handle(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)

	srv := rest.NewServer(t, server.ctxServer.Handler(), nil)

	m := server.newModule("m1", "m1 desc")
	a.NotNil(m)

	path := "/path"
	a.NotError(m.Handle(path, f1, http.MethodGet, http.MethodDelete))
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
	a.NotError(m.Handle(path, f1))
	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusOK)
	srv.NewRequest(http.MethodPatch, path).
		Do().
		Status(http.StatusOK)
}

func TestModule_Handles(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)

	srv := rest.NewServer(t, server.ctxServer.Handler(), nil)

	path := "/path"
	m := server.newModule("m1", "m1 desc")
	a.NotNil(m)

	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusNotFound)

	m.Get(path, f1)
	srv.NewRequest(http.MethodGet, path).
		Do().
		Status(http.StatusOK)

	m.Post(path, f1)
	srv.NewRequest(http.MethodPost, path).
		Do().
		Status(http.StatusOK)

	m.Patch(path, f1)
	srv.NewRequest(http.MethodPatch, path).
		Do().
		Status(http.StatusOK)

	m.Put(path, f1)
	srv.NewRequest(http.MethodPut, path).
		Do().
		Status(http.StatusOK)

	m.Delete(path, f1)
	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusOK)
}
