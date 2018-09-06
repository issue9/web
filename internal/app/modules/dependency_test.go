// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package modules

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/module"
)

var (
	inits   = map[string]int{}
	router  = muxtest.Prefix("")
	infolog = log.New(os.Stderr, "", 0)
	f1      = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("f1"))
		w.WriteHeader(http.StatusAccepted)
	}
)

func i(name string) func() error {
	return func() error {
		inits[name] = inits[name] + 1
		return nil
	}
}

func m(name string, f func() error, deps ...string) *module.Module {
	m := module.New(module.TypeModule, name, name, deps...)
	m.AddInit(f)
	return m
}

func mt(name, title string, f func() error, deps ...string) *module.Module {
	m := module.New(module.TypeModule, name, name, deps...)
	m.AddInitTitle(title, f)
	return m
}

func TestDependency_isDep(t *testing.T) {
	a := assert.New(t)
	dep := newDepencency([]*module.Module{
		m("m1", nil, "d1", "d2"),
		m("d1", nil, "d3"),
	}, router, nil)
	a.NotNil(dep)

	a.True(dep.isDep("m1", "d1"))
	a.True(dep.isDep("m1", "d2"))
	a.True(dep.isDep("m1", "d3")) // 通过 d1 继承
	a.False(dep.isDep("m1", "m1"))

	// 循环依赖
	dep = newDepencency([]*module.Module{
		m("m1", nil, "d1", "d2"),
		m("d1", nil, "d3"),
		m("d3", nil, "d1"),
	}, router, nil)
	a.True(dep.isDep("d1", "d1"))

	// 不存在的模块
	a.False(dep.isDep("d10", "d1"))
}

func TestDependency_checkDeps(t *testing.T) {
	a := assert.New(t)
	dep := newDepencency([]*module.Module{
		m("m1", nil, "d1", "d2"),
		m("d1", nil, "d3"),
	}, router, nil)

	m1 := dep.modules["m1"]
	a.Error(dep.checkDeps(m1)) // 依赖项不存在

	dep = newDepencency([]*module.Module{
		m("m1", nil, "d1", "d2"),
		m("d1", nil, "d3"),
		m("d2", nil, "d3"),
	}, router, nil)
	a.NotError(dep.checkDeps(m1))

	// 自我依赖
	dep = newDepencency([]*module.Module{
		m("m1", nil, "d1", "d2"),
		m("d1", nil, "d3"),
		m("d2", nil, "d3"),
		m("d3", nil, "d2"),
	}, router, nil)
	d2 := dep.modules["d2"]
	a.Error(dep.checkDeps(d2))
}

func TestDependency_init(t *testing.T) {
	a := assert.New(t)

	// 缺少依赖项 d3
	dep := newDepencency([]*module.Module{
		m("m1", i("m1"), "d1", "d2"),
		m("d1", i("d1"), "d3"),
		m("d2", i("d2"), "d3"),
	}, router, infolog)
	a.Error(dep.init(""))

	m1 := m("m1", i("m1"), "d1", "d2")
	m1.PutFunc("/put", f1)
	m1.NewTag("install").PostFunc("/install", f1)
	ms := []*module.Module{
		m1,
		m("d1", i("d1"), "d3"),
		m("d2", i("d2"), "d3"),
		m("d3", i("d3")),
	}

	dep = newDepencency(ms, router, infolog)
	a.NotError(dep.init(""))
	a.Equal(len(inits), 4).
		Equal(inits["m1"], 1).
		Equal(inits["d1"], 1).
		Equal(inits["d2"], 1)

	dep = newDepencency(ms, router, infolog)
	a.NotError(dep.init("install"), infolog)
}
