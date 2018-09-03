// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package modules

import (
	"testing"

	"github.com/issue9/web/module"

	"github.com/issue9/assert"
)

var inits = map[string]int{}

func i(name string) func() error {
	return func() error {
		inits[name] = inits[name] + 1
		return nil
	}
}

func m(name string, f func() error, deps ...string) *module.Module {
	return &module.Module{
		Name:  name,
		Deps:  deps,
		Inits: []*module.Init{&module.Init{F: f}},
	}
}

func TestDependency_isDep(t *testing.T) {
	a := assert.New(t)
	dep := newDepencency([]*module.Module{
		m("m1", i("m1"), "d1", "d2"),
		m("d1", i("d1"), "d3"),
	})
	a.NotNil(dep)

	a.True(dep.isDep("m1", "d1"))
	a.True(dep.isDep("m1", "d2"))
	a.True(dep.isDep("m1", "d3")) // 通过 d1 继承
	a.False(dep.isDep("m1", "m1"))

	// 循环依赖
	dep = newDepencency([]*module.Module{
		m("m1", i("m1"), "d1", "d2"),
		m("d1", i("d1"), "d3"),
		m("d3", i("d3"), "d1"),
	})
	a.True(dep.isDep("d1", "d1"))

	// 不存在的模块
	a.False(dep.isDep("d10", "d1"))
}

func TestDependency_checkDeps(t *testing.T) {
	a := assert.New(t)
	dep := newDepencency([]*module.Module{
		m("m1", i("m1"), "d1", "d2"),
		m("d1", i("d1"), "d3"),
	})

	m1 := dep.modules["m1"]
	a.Error(dep.checkDeps(m1)) // 依赖项不存在

	dep = newDepencency([]*module.Module{
		m("m1", i("m1"), "d1", "d2"),
		m("d1", i("d1"), "d3"),
		m("d2", i("d2"), "d3"),
	})
	a.NotError(dep.checkDeps(m1))

	// 自我依赖
	dep = newDepencency([]*module.Module{
		m("m1", i("m1"), "d1", "d2"),
		m("d1", i("d1"), "d3"),
		m("d2", i("d2"), "d3"),
		m("d3", i("d3"), "d2"),
	})
	d2 := dep.modules["d2"]
	a.Error(dep.checkDeps(d2))
}

func TestDependency_init(t *testing.T) {
	a := assert.New(t)
	dep := newDepencency([]*module.Module{
		m("m1", i("m1"), "d1", "d2"),
		m("d1", i("d1"), "d3"),
		m("d2", i("d2"), "d3"),
	})

	a.Error(dep.init("", router)) // 缺少依赖项 d3

	dep = newDepencency([]*module.Module{
		m("m1", i("m1"), "d1", "d2"),
		m("d1", i("d1"), "d3"),
		m("d2", i("d2"), "d3"),
		m("d3", i("d3")),
	})

	a.NotError(dep.init("", router))
	a.Equal(len(inits), 4).
		Equal(inits["m1"], 1).
		Equal(inits["d1"], 1).
		Equal(inits["d2"], 1)
}
