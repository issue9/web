// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/issue9/assert"
)

func TestTag(t *testing.T) {
	a := assert.New(t)
	m := New(TypeModule, "user1", "user1 desc")
	a.NotNil(m).Equal(m.Type, TypeModule)

	v := m.NewTag("0.1.0")
	a.NotNil(v).NotNil(m.Tags["0.1.0"])
	a.Equal(v.Type, TypeTag).Equal(v.Name, "0.1.0")
	v.AddInit(nil, "title1")
	a.Equal(v.Inits[0].Title, "title1")

	vv := m.NewTag("0.1.0")
	a.Equal(vv, v).Equal(vv.Name, "0.1.0")

	v2 := m.NewTag("0.2.0")
	a.NotEqual(v2, v)

	// 子标签，不能再添加子标
	a.Panic(func() {
		vv.NewTag("0.3.0")
	})
}

func TestModule_AddInit(t *testing.T) {
	a := assert.New(t)

	m := New(TypeModule, "m1", "m1 desc")
	a.NotNil(m)
	m.AddInit(func() error { return nil })
	a.Equal(len(m.Inits), 1).
		NotEmpty(m.Inits[0].Title). // 一个默认的数值。
		NotNil(m.Inits[0].F)

	m.AddInit(func() error { return nil }, "t1")
	a.Equal(len(m.Inits), 2).
		Equal(m.Inits[1].Title, "t1").
		NotNil(m.Inits[1].F)

	m.AddInit(func() error { return nil }, "t1")
	a.Equal(len(m.Inits), 3).
		Equal(m.Inits[2].Title, "t1").
		NotNil(m.Inits[2].F)
}
