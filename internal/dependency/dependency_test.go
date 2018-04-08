// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dependency

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

func TestDependency_New(t *testing.T) {
	a := assert.New(t)
	dep := New()
	a.NotNil(dep)

	a.NotError(dep.Add("m1", i("m1"), "d1", "d2"))
	a.NotError(dep.Add("m2", i("m2"), "d1", "d2"))
	a.Error(dep.Add("m2", i("m2"), "d1", "d2"))
}

func TestDependency_isDep(t *testing.T) {
	a := assert.New(t)
	dep := New()
	a.NotNil(dep)

	a.NotError(dep.Add("m1", i("m1"), "d1", "d2"))
	a.NotError(dep.Add("d1", i("d1"), "d3"))

	a.True(dep.isDep("m1", "d1"))
	a.True(dep.isDep("m1", "d2"))
	a.True(dep.isDep("m1", "d3")) // 通过 d1 继承
	a.False(dep.isDep("m1", "m1"))

	// 循环依赖
	a.NotError(dep.Add("d3", i("d3"), "d1"))
	a.True(dep.isDep("d1", "d1"))

	// 不存在的模块
	a.False(dep.isDep("d10", "d1"))
}

func TestDependency_checkDeps(t *testing.T) {
	a := assert.New(t)
	dep := New()
	a.NotNil(dep)

	a.NotError(dep.Add("m1", i("m1"), "d1", "d2"))
	a.NotError(dep.Add("d1", i("d1"), "d3"))
	m1 := dep.modules["m1"]
	a.Error(dep.checkDeps(m1)) // 依赖项不存在

	a.NotError(dep.Add("d2", i("d2"), "d3"))
	a.NotError(dep.checkDeps(m1))

	// 自我依赖
	d2 := dep.modules["d2"]
	a.NotError(dep.Add("d3", i("d3"), "d2"))
	a.Error(dep.checkDeps(d2))
}

func TestDependency_Init(t *testing.T) {
	a := assert.New(t)
	dep := New()
	a.NotNil(dep)

	a.NotError(dep.Add("m1", i("m1"), "d1", "d2"))
	a.NotError(dep.Add("d1", i("d1"), "d3"))
	a.NotError(dep.Add("d2", i("d2"), "d3"))
	a.Error(dep.Init()) // 缺少依赖项 d3
	a.NotError(dep.Add("d3", i("d3")))

	a.NotError(dep.Init())
	a.Equal(len(inits), 4).
		Equal(inits["m1"], 1).
		Equal(inits["d1"], 1).
		Equal(inits["d2"], 1)
}
