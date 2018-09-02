// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
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

func TestTag(t *testing.T) {
	a := assert.New(t)
	m := New("user1", "user1 desc")
	a.NotNil(m)

	v := m.NewTag("0.1.0")
	a.NotNil(v).NotNil(m.Tags["0.1.0"])
	v.Task("title1", nil)
	a.Equal(v.Inits[0].Title, "title1")
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
