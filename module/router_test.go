// SPDX-License-Identifier: MIT

package module

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
	ms := newModules(a)

	srv := rest.NewServer(t, ms.app.Mux(), nil)

	m := ms.newModule("m1", "m1 desc")
	a.NotNil(m)
	p := m.Prefix("/p")
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
	ms := newModules(a)

	srv := rest.NewServer(t, ms.app.Mux(), nil)

	m := ms.newModule("m1", "m1 desc")
	a.NotNil(m)

	path := "/path"
	a.NotError(m.Handle(path, h1, http.MethodGet, http.MethodDelete))
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
	a.NotError(m.Handle(path, h1))
	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusOK)
	srv.NewRequest(http.MethodPatch, path).
		Do().
		Status(http.StatusOK)
}

func TestModule_Handles(t *testing.T) {
	a := assert.New(t)
	ms := newModules(a)

	srv := rest.NewServer(t, ms.app.Mux(), nil)

	path := "/path"
	m := ms.newModule("m1", "m1 desc")
	a.NotNil(m)

	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusNotFound)

	m.Get(path, h1)
	srv.NewRequest(http.MethodGet, path).
		Do().
		Status(http.StatusOK)

	m.Post(path, h1)
	srv.NewRequest(http.MethodPost, path).
		Do().
		Status(http.StatusOK)

	m.Patch(path, h1)
	srv.NewRequest(http.MethodPatch, path).
		Do().
		Status(http.StatusOK)

	m.Put(path, h1)
	srv.NewRequest(http.MethodPut, path).
		Do().
		Status(http.StatusOK)

	m.Delete(path, h1)
	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusOK)

	// *Func
	path = "/path1"
	m = ms.newModule("m1", "m1 desc")
	a.NotNil(m)

	m.GetFunc(path, f1)
	srv.NewRequest(http.MethodGet, path).
		Do().
		Status(http.StatusOK)

	m.PostFunc(path, f1)
	srv.NewRequest(http.MethodPost, path).
		Do().
		Status(http.StatusOK)

	m.PatchFunc(path, f1)
	srv.NewRequest(http.MethodPatch, path).
		Do().
		Status(http.StatusOK)

	m.PutFunc(path, f1)
	srv.NewRequest(http.MethodPut, path).
		Do().
		Status(http.StatusOK)

	m.DeleteFunc(path, f1)
	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusOK)
}
