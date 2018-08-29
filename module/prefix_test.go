// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/issue9/assert"
)

func TestPrefix_Module(t *testing.T) {
	a := assert.New(t)

	m := New(router, "m1", "m1 desc")
	a.NotNil(m)

	p := m.Prefix("/p")
	a.Equal(p.Module(), m)
}

func TestPrefix_Handles(t *testing.T) {
	a := assert.New(t)

	path := "/path"
	m := New(router, "m1", "m1 desc")
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
	m = New(router, "m1", "m1 desc")
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
