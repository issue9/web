// SPDX-License-Identifier: MIT

package web

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/context"
)

func m(web *Web, name string, f func() error, deps ...string) *Module {
	m := web.NewModule(name, name, deps...)
	m.AddInit(f, "init")
	return m
}

func newDepWeb(ms []*Module) *Web {
	return &Web{modules: ms}
}

func TestDependency_isDep(t *testing.T) {
	a := assert.New(t)
	srv := newWeb(a)

	dep := newDepWeb([]*Module{
		m(srv, "m1", nil, "d1", "d2"),
		m(srv, "d1", nil, "d3"),
	})
	a.NotNil(dep)

	a.True(dep.isDep("m1", "d1"))
	a.True(dep.isDep("m1", "d2"))
	a.True(dep.isDep("m1", "d3")) // 通过 d1 继承
	a.False(dep.isDep("m1", "m1"))

	// 循环依赖
	dep = newDepWeb([]*Module{
		m(srv, "m1", nil, "d1", "d2"),
		m(srv, "d1", nil, "d3"),
		m(srv, "d3", nil, "d1"),
	})
	a.True(dep.isDep("d1", "d1"))

	// 不存在的模块
	a.False(dep.isDep("d10", "d1"))
}

func TestDependency_checkDeps(t *testing.T) {
	a := assert.New(t)
	srv := newWeb(a)

	dep := newDepWeb([]*Module{
		m(srv, "m1", nil, "d1", "d2"),
		m(srv, "d1", nil, "d3"),
	})

	m1 := dep.module("m1")
	a.NotNil(m1).
		Error(dep.checkDeps(m1)) // 依赖项不存在

	dep = newDepWeb([]*Module{
		m(srv, "m1", nil, "d1", "d2"),
		m(srv, "d1", nil, "d3"),
		m(srv, "d2", nil, "d3"),
	})
	a.NotError(dep.checkDeps(m1))

	// 自我依赖
	dep = newDepWeb([]*Module{
		m(srv, "m1", nil, "d1", "d2"),
		m(srv, "d1", nil, "d3"),
		m(srv, "d2", nil, "d3"),
		m(srv, "d3", nil, "d2"),
	})
	d2 := dep.module("d2")
	a.NotNil(d2).
		Error(dep.checkDeps(d2))
}

func TestDependency_init(t *testing.T) {
	a := assert.New(t)
	srv := newWeb(a)

	inits := map[string]int{}
	infolog := log.New(os.Stderr, "", 0)
	f1 := func(ctx *context.Context) {
		ctx.Render(http.StatusAccepted, "f1", nil)
	}
	i := func(name string) func() error {
		return func() error {
			inits[name] = inits[name] + 1
			return nil
		}
	}

	// 缺少依赖项 d3
	dep := newDepWeb([]*Module{
		m(srv, "m1", i("m1"), "d1", "d2"),
		m(srv, "d1", i("d1"), "d3"),
		m(srv, "d2", i("d2"), "d3"),
	})
	a.Error(dep.initDeps("", infolog))

	m1 := m(srv, "m1", i("m1"), "d1", "d2")
	m1.Put("/put", f1)
	apps := []*Module{
		m1,
		m(srv, "d1", i("d1"), "d3"),
		m(srv, "d2", i("d2"), "d3"),
		m(srv, "d3", i("d3")),
	}

	dep = newDepWeb(apps)
	a.NotError(dep.initDeps("", infolog))
	a.Equal(len(inits), 4).
		Equal(inits["m1"], 1).
		Equal(inits["d1"], 1).
		Equal(inits["d2"], 1)

	dep = newDepWeb(apps)
	a.NotError(dep.initDeps("install", infolog))
}
