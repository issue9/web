// SPDX-License-Identifier: MIT

package dep

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/issue9/assert"
)

func newDep(a *assert.Assertion, ms []Module) *Dep {
	d := New(log.New(ioutil.Discard, "", 0))
	a.NotNil(d)

	for _, m := range ms {
		a.NotError(d.AddModule(m))
	}

	a.Equal(len(ms), len(d.Modules()))

	return d
}

func TestDep_isDep(t *testing.T) {
	a := assert.New(t)

	d := newDep(a, []Module{
		newMod("m1", nil, "d1", "d2"),
		newMod("d1", nil, "d3"),
	})

	a.True(d.isDep("m1", "d1"))
	a.True(d.isDep("m1", "d2"))
	a.True(d.isDep("m1", "d3")) // 通过 d1 继承
	a.False(d.isDep("m1", "m1"))

	// 循环依赖
	d = newDep(a, []Module{
		newMod("m1", nil, "d1", "d2"),
		newMod("d1", nil, "d3"),
		newMod("d3", nil, "d1"),
	})
	a.True(d.isDep("d1", "d1"))

	// 不存在的模块
	a.False(d.isDep("d10", "d1"))
}

func TestDep_checkDeps(t *testing.T) {
	a := assert.New(t)

	d := newDep(a, []Module{
		newMod("m1", nil, "d1", "d2"),
		newMod("d1", nil, "d3"),
	})

	m1 := d.findModule("m1")
	a.NotNil(m1).
		Error(d.checkDeps(m1)) // 依赖项不存在

	d = newDep(a, []Module{
		newMod("m1", nil, "d1", "d2"),
		newMod("d1", nil, "d3"),
		newMod("d2", nil, "d3"),
	})
	a.NotError(d.checkDeps(m1))

	// 自我依赖
	d = newDep(a, []Module{
		newMod("m1", nil, "d1", "d2"),
		newMod("d1", nil, "d3"),
		newMod("d2", nil, "d3"),
		newMod("d3", nil, "d2"),
	})
	d2 := d.findModule("d2")
	a.NotNil(d2).
		Error(d.checkDeps(d2))
}

func TestDep_Init(t *testing.T) {
	a := assert.New(t)

	inits := map[string]int{}

	f := func(name string) func() error {
		return func() error {
			inits[name] = inits[name] + 1
			return nil
		}
	}

	// 缺少依赖项 d3
	d := newDep(a, []Module{
		newMod("m1", f("m1"), "d1", "d2"),
		newMod("d1", f("d1"), "d3"),
		newMod("d2", f("d2"), "d3"),
	})
	a.Error(d.Init())

	d = newDep(a, []Module{
		newMod("m1", f("m1"), "d1", "d2"),
		newMod("d1", f("d1"), "d3"),
		newMod("d2", f("d2"), "d3"),
		newMod("d3", f("d3")),
	})

	a.NotError(d.Init())
	a.Equal(len(inits), 4).
		Equal(inits["m1"], 1). // 为 1 表示不会被多次调用
		Equal(inits["d1"], 1).
		Equal(inits["d2"], 1).
		Equal(inits["d3"], 1)

	a.ErrorIs(d.Init(), ErrInited)

	// 添加已经存在的模块
	a.ErrorIs(d.AddModule(newMod("m1", f("m1"), "d1")), ErrModExists)

	// 添加新模块，会自动调用初始化函数
	a.NotError(d.AddModule(newMod("d4", f("d4"), "d1")))
	a.Equal(len(inits), 5).
		Equal(inits["m1"], 1). // 为 1 表示不会被多次调用
		Equal(inits["d1"], 1).
		Equal(inits["d2"], 1).
		Equal(inits["d3"], 1).
		Equal(inits["d4"], 1)
}
