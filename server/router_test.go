// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

var (
	f1 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	h1 = http.HandlerFunc(f1)
)

func TestModule_Prefix(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)
	srv := rest.NewServer(t, app.Mux(), nil)

	p := app.Prefix("/p")
	a.NotNil(p)

	path := "/path"
	a.NotError(p.Handle(path, h1, http.MethodGet, http.MethodDelete))
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
	app := newApp(a)
	srv := rest.NewServer(t, app.Mux(), nil)

	path := "/path"
	a.NotError(app.Handle(path, h1, http.MethodGet, http.MethodDelete))
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
	a.NotError(app.Handle(path, h1))
	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusOK)
	srv.NewRequest(http.MethodPatch, path).
		Do().
		Status(http.StatusOK)
}

func TestModule_Handles(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)
	srv := rest.NewServer(t, app.Mux(), nil)

	path := "/path"

	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusNotFound)

	app.Get(path, h1)
	srv.NewRequest(http.MethodGet, path).
		Do().
		Status(http.StatusOK)

	app.Post(path, h1)
	srv.NewRequest(http.MethodPost, path).
		Do().
		Status(http.StatusOK)

	app.Patch(path, h1)
	srv.NewRequest(http.MethodPatch, path).
		Do().
		Status(http.StatusOK)

	app.Put(path, h1)
	srv.NewRequest(http.MethodPut, path).
		Do().
		Status(http.StatusOK)

	app.Delete(path, h1)
	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusOK)

	// *Func
	path = "/path1"

	app.GetFunc(path, f1)
	srv.NewRequest(http.MethodGet, path).
		Do().
		Status(http.StatusOK)

	app.PostFunc(path, f1)
	srv.NewRequest(http.MethodPost, path).
		Do().
		Status(http.StatusOK)

	app.PatchFunc(path, f1)
	srv.NewRequest(http.MethodPatch, path).
		Do().
		Status(http.StatusOK)

	app.PutFunc(path, f1)
	srv.NewRequest(http.MethodPut, path).
		Do().
		Status(http.StatusOK)

	app.DeleteFunc(path, f1)
	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusOK)
}
