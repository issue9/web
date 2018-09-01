// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux"
)

var (
	f1 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	h1 = http.HandlerFunc(f1)

	router = mux.New(false, false, nil, nil).Prefix("")
)

func TestModule_GetInit(t *testing.T) {
	a := assert.New(t)

	m := New("m1", "m1 desc")
	a.NotNil(m)
	fn := m.GetInit(router)
	a.NotNil(fn).NotError(fn())

	// 返回错误
	m = New("m2", "m2 desc")
	a.NotNil(m)
	m.AddInit(func() error {
		return errors.New("error")
	})
	fn = m.GetInit(router)
	a.NotNil(fn).ErrorString(fn(), "error")

	w := new(bytes.Buffer)
	m = New("m3", "m3 desc")
	a.NotNil(m)
	m.AddInit(func() error {
		_, err := w.WriteString("m3")
		return err
	})
	m.GetFunc("/get", f1)
	m.Prefix("/p").PostFunc("/post", f1)
	fn = m.GetInit(router)
	a.NotNil(fn).
		NotError(fn()).
		Equal(w.String(), "m3")
}

func TestModule_Handles(t *testing.T) {
	a := assert.New(t)

	path := "/path"
	m := New("m1", "m1 desc")
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
	m = New("m1", "m1 desc")
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
