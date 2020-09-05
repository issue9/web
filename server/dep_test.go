// SPDX-License-Identifier: MIT

package server

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/issue9/assert"
)

func m(srv *Server, name string, f func() error, deps ...string) *Module {
	m := srv.newModule(name, name, deps...)
	m.AddInit(f, "init")
	return m
}

func newDep(ms []*Module, log *log.Logger) *dependency {
	return newDepencency(ms, log)
}

func TestDependency_isDep(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

	dep := newDep([]*Module{
		m(srv, "m1", nil, "d1", "d2"),
		m(srv, "d1", nil, "d3"),
	}, nil)
	a.NotNil(dep)

	a.True(dep.isDep("m1", "d1"))
	a.True(dep.isDep("m1", "d2"))
	a.True(dep.isDep("m1", "d3")) // 通过 d1 继承
	a.False(dep.isDep("m1", "m1"))

	// 循环依赖
	dep = newDep([]*Module{
		m(srv, "m1", nil, "d1", "d2"),
		m(srv, "d1", nil, "d3"),
		m(srv, "d3", nil, "d1"),
	}, nil)
	a.True(dep.isDep("d1", "d1"))

	// 不存在的模块
	a.False(dep.isDep("d10", "d1"))
}

func TestDependency_checkDeps(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

	dep := newDep([]*Module{
		m(srv, "m1", nil, "d1", "d2"),
		m(srv, "d1", nil, "d3"),
	}, nil)

	m1 := dep.modules["m1"]
	a.Error(dep.checkDeps(m1)) // 依赖项不存在

	dep = newDep([]*Module{
		m(srv, "m1", nil, "d1", "d2"),
		m(srv, "d1", nil, "d3"),
		m(srv, "d2", nil, "d3"),
	}, nil)
	a.NotError(dep.checkDeps(m1))

	// 自我依赖
	dep = newDep([]*Module{
		m(srv, "m1", nil, "d1", "d2"),
		m(srv, "d1", nil, "d3"),
		m(srv, "d2", nil, "d3"),
		m(srv, "d3", nil, "d2"),
	}, nil)
	d2 := dep.modules["d2"]
	a.Error(dep.checkDeps(d2))
}

func TestDependency_init(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

	inits := map[string]int{}
	infolog := log.New(os.Stderr, "", 0)
	f1 := func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("f1"))
		a.NotError(err)
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
		m(srv, "m1", i("m1"), "d1", "d2"),
		m(srv, "d1", i("d1"), "d3"),
		m(srv, "d2", i("d2"), "d3"),
	}, infolog)
	a.Error(dep.init(""))

	m1 := m(srv, "m1", i("m1"), "d1", "d2")
	m1.PutFunc("/put", f1)
	apps := []*Module{
		m1,
		m(srv, "d1", i("d1"), "d3"),
		m(srv, "d2", i("d2"), "d3"),
		m(srv, "d3", i("d3")),
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
