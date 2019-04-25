// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/internal/webconfig"
)

func TestModule_NewTag(t *testing.T) {
	a := assert.New(t)
	ms, err := NewModules(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)
	m := newModule(ms, "user1", "user1 desc")
	a.NotNil(m)

	v := m.NewTag("0.1.0")
	a.NotNil(v).NotNil(m.tags["0.1.0"])
	v.AddInit(nil, "title1")
	a.Equal(v.inits[0].title, "title1")

	vv := m.NewTag("0.1.0")
	a.Equal(vv, v)

	v2 := m.NewTag("0.2.0")
	a.NotEqual(v2, v)
}

func TestModule_Plugin(t *testing.T) {
	a := assert.New(t)

	ms, err := NewModules(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)
	m := newModule(ms, "user1", "user1 desc")
	a.NotNil(m)

	a.Panic(func() {
		m.Plugin("p1", "p1 desc")
	})

	m = newModule(ms, "", "")
	a.NotPanic(func() {
		m.Plugin("p1", "p1 desc")
	})
}

func TestModule_AddInit(t *testing.T) {
	a := assert.New(t)
	ms, err := NewModules(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)
	m := newModule(ms, "m1", "m1 desc")
	a.NotNil(m)

	a.Nil(m.inits)
	m.AddInit(func() error { return nil }, "t1")
	a.Equal(len(m.inits), 1).
		Equal(m.inits[0].title, "t1").
		NotNil(m.inits[0].f)

	m.AddInit(func() error { return nil }, "t1")
	a.Equal(len(m.inits), 2).
		Equal(m.inits[1].title, "t1").
		NotNil(m.inits[1].f)

	m.AddInit(func() error { return nil }, "t1")
	a.Equal(len(m.inits), 3).
		Equal(m.inits[2].title, "t1").
		NotNil(m.inits[2].f)
}
