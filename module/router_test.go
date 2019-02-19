// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

var (
	f1 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	h1 = http.HandlerFunc(f1)
)

func TestPrefix_Module(t *testing.T) {
	a := assert.New(t)

	m := newModule(TypeModule, "m1", "m1 desc")
	a.NotNil(m)

	p := m.Prefix("/p")
	a.Equal(p.Module(), m)
}

func TestModule_Handle(t *testing.T) {
	a := assert.New(t)

	m := newModule(TypeModule, "m1", "m1 desc")
	a.NotNil(m)

	path := "/path"
	m.Handle(path, h1, http.MethodGet, http.MethodDelete)
	a.Equal(len(m.Routes[path]), 2)

	path = "/path1"
	m.Handle(path, h1)
	a.Equal(len(m.Routes[path]), len(defaultMethods))
}

func TestModule_Handles(t *testing.T) {
	a := assert.New(t)

	path := "/path"
	m := newModule(TypeModule, "m1", "m1 desc")
	a.NotNil(m)

	m.Get(path, h1)
	a.Panic(func() { m.GetFunc(path, f1) })
	a.Equal(len(m.Routes[path]), 1)

	m.Post(path, h1)
	a.Equal(len(m.Routes[path]), 2)

	m.Patch(path, h1)
	a.Equal(len(m.Routes[path]), 3)

	m.Put(path, h1)
	a.Equal(len(m.Routes[path]), 4)

	m.Delete(path, h1)
	a.Equal(len(m.Routes[path]), 5)

	// *Func
	path = "/path1"
	m = newModule(TypeModule, "m1", "m1 desc")
	a.NotNil(m)

	m.GetFunc(path, f1)
	a.Equal(len(m.Routes[path]), 1)

	m.PostFunc(path, f1)
	a.Equal(len(m.Routes[path]), 2)

	m.PatchFunc(path, f1)
	a.Equal(len(m.Routes[path]), 3)

	m.PutFunc(path, f1)
	a.Equal(len(m.Routes[path]), 4)

	m.DeleteFunc(path, f1)
	a.Equal(len(m.Routes[path]), 5)
}

func TestPrefix_Handles(t *testing.T) {
	a := assert.New(t)

	path := "/path"
	m := newModule(TypeModule, "m1", "m1 desc")
	a.NotNil(m)
	p := m.Prefix("/p")
	mp := p.prefix + path

	p.Get(path, h1)
	a.Panic(func() { p.GetFunc(path, f1) })
	a.Equal(len(m.Routes[mp]), 1)

	p.Post(path, h1)
	a.Equal(len(m.Routes[mp]), 2)

	p.Patch(path, h1)
	a.Equal(len(m.Routes[mp]), 3)

	p.Put(path, h1)
	a.Equal(len(m.Routes[mp]), 4)

	p.Delete(path, h1)
	a.Equal(len(m.Routes[mp]), 5)

	// *Func
	path = "/path1"
	m = newModule(TypeModule, "m1", "m1 desc")
	a.NotNil(m)
	p = m.Prefix("/p")
	mp = p.prefix + path

	p.GetFunc(path, f1)
	a.Equal(len(m.Routes[mp]), 1)

	p.PostFunc(path, f1)
	a.Equal(len(m.Routes[mp]), 2)

	p.PatchFunc(path, f1)
	a.Equal(len(m.Routes[mp]), 3)

	p.PutFunc(path, f1)
	a.Equal(len(m.Routes[mp]), 4)

	p.DeleteFunc(path, f1)
	a.Equal(len(m.Routes[mp]), 5)
}
