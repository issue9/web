// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/issue9/assert"
)

func m(ms *Modules, name string, f func() error, deps ...string) *Module {
	m := ms.newModule(name, name, deps...)
	m.AddInit(f, "init")
	return m
}

func newDep(ms []*Module, log *log.Logger) *dependency {
	return newDepencency(ms, log)
}

func TestDependency_isDep(t *testing.T) {
	a := assert.New(t)
	ms := newModules(a)

	dep := newDep([]*Module{
		m(ms, "m1", nil, "d1", "d2"),
		m(ms, "d1", nil, "d3"),
	}, nil)
	a.NotNil(dep)

	a.True(dep.isDep("m1", "d1"))
	a.True(dep.isDep("m1", "d2"))
	a.True(dep.isDep("m1", "d3")) // 通过 d1 继承
	a.False(dep.isDep("m1", "m1"))

	// 循环依赖
	dep = newDep([]*Module{
		m(ms, "m1", nil, "d1", "d2"),
		m(ms, "d1", nil, "d3"),
		m(ms, "d3", nil, "d1"),
	}, nil)
	a.True(dep.isDep("d1", "d1"))

	// 不存在的模块
	a.False(dep.isDep("d10", "d1"))
}

func TestDependency_checkDeps(t *testing.T) {
	a := assert.New(t)
	ms := newModules(a)

	dep := newDep([]*Module{
		m(ms, "m1", nil, "d1", "d2"),
		m(ms, "d1", nil, "d3"),
	}, nil)

	m1 := dep.modules["m1"]
	a.Error(dep.checkDeps(m1)) // 依赖项不存在

	dep = newDep([]*Module{
		m(ms, "m1", nil, "d1", "d2"),
		m(ms, "d1", nil, "d3"),
		m(ms, "d2", nil, "d3"),
	}, nil)
	a.NotError(dep.checkDeps(m1))

	// 自我依赖
	dep = newDep([]*Module{
		m(ms, "m1", nil, "d1", "d2"),
		m(ms, "d1", nil, "d3"),
		m(ms, "d2", nil, "d3"),
		m(ms, "d3", nil, "d2"),
	}, nil)
	d2 := dep.modules["d2"]
	a.Error(dep.checkDeps(d2))
}

func TestDependency_init(t *testing.T) {
	a := assert.New(t)
	ms := newModules(a)

	inits := map[string]int{}
	infolog := log.New(os.Stderr, "", 0)
	f1 := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("f1"))
		w.WriteHeader(http.StatusAccepted)
	}
	i := func(name string) func() error {
		return func() error {
			inits[name] = inits[name] + 1
			return nil
		}
	}

	// 缺少依赖项 d3
	dep := newDep([]*Module{
		m(ms, "m1", i("m1"), "d1", "d2"),
		m(ms, "d1", i("d1"), "d3"),
		m(ms, "d2", i("d2"), "d3"),
	}, infolog)
	a.Error(dep.init(""))

	m1 := m(ms, "m1", i("m1"), "d1", "d2")
	m1.PutFunc("/put", f1)
	apps := []*Module{
		m1,
		m(ms, "d1", i("d1"), "d3"),
		m(ms, "d2", i("d2"), "d3"),
		m(ms, "d3", i("d3")),
	}

	dep = newDep(apps, infolog)
	a.NotError(dep.init(""))
	a.Equal(len(inits), 4).
		Equal(inits["m1"], 1).
		Equal(inits["d1"], 1).
		Equal(inits["d2"], 1)

	dep = newDep(apps, infolog)
	a.NotError(dep.init("install"), infolog)
}
