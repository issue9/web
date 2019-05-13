// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

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
	srv := rest.NewServer(t, Mux(), nil)
	Mux().Clean()

	m := newModule("m1", "m1 desc")
	a.NotNil(m)
	p := m.Prefix("/p")
	a.NotNil(p)

	path := "/path"
	p.Handle(path, h1, http.MethodGet, http.MethodDelete)
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
	srv := rest.NewServer(t, Mux(), nil)
	Mux().Clean()

	m := newModule("m1", "m1 desc")
	a.NotNil(m)

	path := "/path"
	m.Handle(path, h1, http.MethodGet, http.MethodDelete)
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
	m.Handle(path, h1)
	srv.NewRequest(http.MethodDelete, path).
		Do().
		Status(http.StatusOK)
	srv.NewRequest(http.MethodPatch, path).
		Do().
		Status(http.StatusOK)
}

func TestModule_Handles(t *testing.T) {
	a := assert.New(t)
	srv := rest.NewServer(t, Mux(), nil)
	Mux().Clean()

	path := "/path"
	m := newModule("m1", "m1 desc")
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
	m = newModule("m1", "m1 desc")
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
