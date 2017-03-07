// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package modules

import (
	"testing"

	"github.com/issue9/assert"
)

var inits = map[string]int{}

func i(name string) func() error {
	return func() error {
		inits[name] = inits[name] + 1
		return nil
	}
}

func TestModules_New(t *testing.T) {
	a := assert.New(t)
	ms := New()
	a.NotNil(ms)

	a.NotError(ms.New("m1", i("m1"), "d1", "d2"))
	a.NotError(ms.New("m2", i("m2"), "d1", "d2"))
	a.Error(ms.New("m2", i("m2"), "d1", "d2"))
}

func TestModules_isDep(t *testing.T) {
	a := assert.New(t)
	ms := New()
	a.NotNil(ms)

	a.NotError(ms.New("m1", i("m1"), "d1", "d2"))
	a.NotError(ms.New("d1", i("d1"), "d3"))

	a.True(ms.isDep("m1", "d1"))
	a.True(ms.isDep("m1", "d2"))
	a.True(ms.isDep("m1", "d3")) // 通过 d1 继承
	a.False(ms.isDep("m1", "m1"))

	// 循环依赖
	a.NotError(ms.New("d3", i("d3"), "d1"))
	a.True(ms.isDep("d1", "d1"))

	// 不存在的模块
	a.False(ms.isDep("d10", "d1"))
}

func TestModules_checkDeps(t *testing.T) {
	a := assert.New(t)
	ms := New()
	a.NotNil(ms)

	a.NotError(ms.New("m1", i("m1"), "d1", "d2"))
	a.NotError(ms.New("d1", i("d1"), "d3"))
	m1 := ms.modules["m1"]
	a.Error(ms.checkDeps(m1)) // 依赖项不存在

	a.NotError(ms.New("d2", i("d2"), "d3"))
	a.NotError(ms.checkDeps(m1))
}

func TestModules_Init(t *testing.T) {
	a := assert.New(t)
	ms := New()
	a.NotNil(ms)

	a.NotError(ms.New("m1", i("m1"), "d1", "d2"))
	a.NotError(ms.New("d1", i("d1"), "d3"))
	a.NotError(ms.New("d2", i("d2"), "d3"))
	a.NotError(ms.New("d3", i("d3")))

	a.NotError(ms.Init())
	a.Equal(len(inits), 4).
		Equal(inits["m1"], 1).
		Equal(inits["d1"], 1).
		Equal(inits["d2"], 1)
}
