// SPDX-License-Identifier: MIT

package module

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v2"
)

func newDep(a *assert.Assertion, ms []*Module) *Dep {
	d := NewDep(logs.New())
	a.NotNil(d)

	for _, m := range ms {
		a.NotError(d.Add(m))
	}

	a.Equal(len(ms), len(d.Modules()))

	return d
}

func TestDep_Tags(t *testing.T) {
	a := assert.New(t)
	d := NewDep(logs.New())
	a.Empty(d.Tags())

	m1 := NewModule("m1", "m1 desc")
	a.NotNil(m1.GetTag("d1"))
	a.NotNil(m1.GetTag("d2"))
	a.NotNil(m1.GetTag("d2"))
	a.NotError(d.Add(m1))

	m2 := NewModule("m2", "m2 desc")
	a.NotNil(m2.GetTag("d2"))
	a.NotNil(m2.GetTag("d0"))
	a.NotError(d.Add(m2))

	m3 := NewModule("m3", "m3 desc")
	a.NotError(d.Add(m3))

	a.Equal(d.Tags(), []string{"d0", "d1", "d2"})
}

func TestDep_InitTag(t *testing.T) {
	a := assert.New(t)
	d := NewDep(logs.New())
	m1 := NewModule("m1", "m1 desc")
	a.NotNil(m1.GetTag("d1"))
	a.NotNil(m1.GetTag("d2"))
	a.NotError(d.Add(m1))

	d.Init("d1")
	a.ErrorType(d.Init("d1"), ErrInited)
}

func TestDep_isDep(t *testing.T) {
	a := assert.New(t)

	d := newDep(a, []*Module{
		NewModule("m1", "m1 desc", "d1", "d2"),
		NewModule("d1", "d1 desc", "d3"),
	})

	a.True(d.isDep("m1", "d1"))
	a.True(d.isDep("m1", "d2"))
	a.True(d.isDep("m1", "d3")) // 通过 d1 继承
	a.False(d.isDep("m1", "m1"))

	// 循环依赖
	d = newDep(a, []*Module{
		NewModule("m1", "m1 desc", "d1", "d2"),
		NewModule("d1", "d1 desc", "d3"),
		NewModule("d3", "d3 desc", "d1"),
	})
	a.True(d.isDep("d1", "d1"))

	// 不存在的模块
	a.False(d.isDep("d10", "d1"))
}

func TestDep_checkDeps(t *testing.T) {
	a := assert.New(t)

	d := newDep(a, []*Module{
		NewModule("m1", "m1 desc", "d1", "d2"),
		NewModule("d1", "d1 desc", "d3"),
	})

	m1 := d.findModule("m1")
	a.NotNil(m1).
		Error(d.checkDeps(m1)) // 依赖项不存在

	d = newDep(a, []*Module{
		NewModule("m1", "m1 desc", "d1", "d2"),
		NewModule("d1", "d1 desc", "d3"),
		NewModule("d2", "d2 desc", "d3"),
	})
	a.NotError(d.checkDeps(m1))

	// 自我依赖
	d = newDep(a, []*Module{
		NewModule("m1", "m1 desc", "d1", "d2"),
		NewModule("d1", "d1 desc", "d3"),
		NewModule("d2", "d2 desc", "d3"),
		NewModule("d3", "d3 desc", "d2"),
	})
	d2 := d.findModule("d2")
	a.NotNil(d2).
		Error(d.checkDeps(d2))
}

func TestDep_Init(t *testing.T) {
	a := assert.New(t)

	inits := map[string]int{}

	newMod := func(name string, dep ...string) *Module {
		m := NewModule(name, name+" desc", dep...)
		m.AddInit("init "+name, func() error {
			inits[name] = inits[name] + 1
			return nil
		})
		return m
	}

	// 缺少依赖项 d3
	d := newDep(a, []*Module{
		newMod("m1", "d1", "d2"),
		newMod("d1", "d3"),
		newMod("d2", "d3"),
	})
	a.Error(d.Init(""))

	d = newDep(a, []*Module{
		newMod("m1", "d1", "d2"),
		newMod("d1", "d3"),
		newMod("d2", "d3"),
		newMod("d3"),
	})

	a.NotError(d.Init(""))
	a.Equal(len(inits), 4).
		Equal(inits["m1"], 1). // 为 1 表示不会被多次调用
		Equal(inits["d1"], 1).
		Equal(inits["d2"], 1).
		Equal(inits["d3"], 1)

	a.ErrorType(d.Init(""), ErrInited)

	// 添加已经存在的模块
	a.ErrorString(d.Add(newMod("m1", "d1")), "已经存在")

	// 添加新模块，会自动调用初始化函数
	a.NotError(d.Add(newMod("d4", "d1")))
	a.Equal(len(inits), 5).
		Equal(inits["m1"], 1). // 为 1 表示不会被多次调用
		Equal(inits["d1"], 1).
		Equal(inits["d2"], 1).
		Equal(inits["d3"], 1).
		Equal(inits["d4"], 1)
}
