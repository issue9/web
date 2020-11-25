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

func TestDep_NewItem(t *testing.T) {
	a := assert.New(t)
	d := New(log.New(ioutil.Discard, "", 0))

	d1 := d.NewItem("d1")
	d2 := d.NewItem("d2")
	a.False(d1 == d2) // 指向不同的内在
	d11 := d.NewItem("d1")
	a.True(d1 == d11) // 指向相同的地址
}

func TestDep_Items(t *testing.T) {
	a := assert.New(t)
	d := New(log.New(ioutil.Discard, "", 0))

	d1 := d.NewItem("d1")
	d2 := d.NewItem("d2")

	a.Empty(d.Items())
	a.Empty(d.Items("d1"))

	a.NotError(d1.AddModule(NewDefaultModule("m1", "m1 desc")))
	a.NotError(d2.AddModule(NewDefaultModule("m2", "m2 desc")))
	a.NotError(d2.AddModule(NewDefaultModule("m1", "m1 desc")))
	a.Equal(d.Items(), map[string][]string{"m1": {"d1", "d2"}, "m2": {"d2"}})
	a.Equal(d.Items("m2"), map[string][]string{"m2": {"d2"}})
}

func TestDep_InitItem(t *testing.T) {
	a := assert.New(t)
	d := New(log.New(ioutil.Discard, "", 0))
	d1 := d.NewItem("d1")
	d2 := d.NewItem("d2")

	a.Panic(func() {
		d.InitItem("")
	})

	a.False(d1.inited).False(d2.inited)
	d.InitItem("d1")
	a.True(d1.inited).False(d2.inited)

	a.Error(d.InitItem("exists"))
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

	m1 := d.FindModule("m1")
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
	d2 := d.FindModule("d2")
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

	a.Panic(func() {
		d.Init()
	})

	// 添加已经存在的模块
	a.ErrorString(d.AddModule(newMod("m1", f("m1"), "d1")), "已经存在")

	// 添加新模块，会自动调用初始化函数
	a.NotError(d.AddModule(newMod("d4", f("d4"), "d1")))
	a.Equal(len(inits), 5).
		Equal(inits["m1"], 1). // 为 1 表示不会被多次调用
		Equal(inits["d1"], 1).
		Equal(inits["d2"], 1).
		Equal(inits["d3"], 1).
		Equal(inits["d4"], 1)
}