// SPDX-License-Identifier: MIT

package dep

import (
	"log"
	"testing"

	"github.com/issue9/assert"
)

func TestDep_Tags(t *testing.T) {
	a := assert.New(t)
	d := NewDep()
	a.Empty(d.Tags())

	m1, err := d.NewModule("m1", "m1 desc")
	a.NotError(err).NotNil(m1)
	a.NotNil(m1.Tag("d1"))
	a.NotNil(m1.Tag("d2"))
	a.NotNil(m1.Tag("d2"))

	m2, err := d.NewModule("m2", "m2 desc")
	a.NotError(err).NotNil(m2)
	a.NotNil(m2.Tag("d2"))
	a.NotNil(m2.Tag("d0"))

	m3, err := d.NewModule("m3", "m3 desc")
	a.NotError(err).NotNil(m3)

	a.Equal(d.Tags(), []string{"d0", "d1", "d2"}).
		Equal(d.Modules(), []*Module{m1, m2, m3})
}

func TestDep_isDep(t *testing.T) {
	a := assert.New(t)
	d := NewDep()
	a.Empty(d.Tags())

	m, err := d.NewModule("m1", "m1 desc", "d1", "d2")
	a.NotError(err).NotNil(m)
	m, err = d.NewModule("d1", "d1 desc", "d3")
	a.NotError(err).NotNil(m)

	a.True(d.isDep("m1", "d1"))
	a.True(d.isDep("m1", "d2"))
	a.True(d.isDep("m1", "d3")) // 通过 d1 继承
	a.False(d.isDep("m1", "m1"))

	// 循环依赖
	d = NewDep()
	a.Empty(d.Tags())
	m, err = d.NewModule("m1", "m1 desc", "d1", "d2")
	a.NotError(err).NotNil(m)
	m, err = d.NewModule("d1", "d1 desc", "d3")
	a.NotError(err).NotNil(m)
	m, err = d.NewModule("d3", "d3 desc", "d1")
	a.NotError(err).NotNil(m)
	a.True(d.isDep("d1", "d1"))

	// 不存在的模块
	a.False(d.isDep("d10", "d1"))
}

func TestDep_checkDeps(t *testing.T) {
	a := assert.New(t)
	d := NewDep()
	a.Empty(d.Tags())

	d.NewModule("m1", "m1 desc", "d1", "d2")
	d.NewModule("d1", "d1 desc", "d3")

	m1 := d.findModule("m1")
	a.NotNil(m1).
		ErrorString(d.checkDeps(m1), "未找到") // 依赖项不存在

	d = NewDep()
	a.NotNil(d)
	d.NewModule("m1", "m1 desc", "d1", "d2")
	d.NewModule("d1", "d1 desc", "d3")
	d.NewModule("d2", "d2 desc", "d3")
	a.NotError(d.checkDeps(m1))

	// 自我依赖
	d = NewDep()
	a.NotNil(d)
	d.NewModule("m1", "m1 desc", "d1", "d2")
	d.NewModule("d1", "d1 desc", "d3")
	d.NewModule("d2", "d2 desc", "d3")
	d.NewModule("d3", "d3 desc", "d2")
	d2 := d.findModule("d2")
	a.NotNil(d2).
		ErrorString(d.checkDeps(d2), "循环依赖自身")
}

func TestDep_Init(t *testing.T) {
	a := assert.New(t)

	inits := map[string]int{}

	newMod := func(d *Dep, name string, dep ...string) {
		m, err := d.NewModule(name, name+" desc", dep...)
		a.NotError(err).NotNil(m)
		m.Tag("default").On("init "+name, func() error {
			inits[name] = inits[name] + 1
			return nil
		})
	}

	// 缺少依赖项 d3
	d := NewDep()
	a.NotNil(d)
	newMod(d, "m1", "d1", "d2")
	newMod(d, "d1", "d3")
	newMod(d, "d2", "d3")
	a.ErrorString(d.Init(log.Default(), "default"), "未找到")

	d = NewDep()
	a.NotNil(d)
	inits = map[string]int{}
	newMod(d, "m1", "d1", "d2")
	newMod(d, "d1", "d3")
	newMod(d, "d2", "d3")
	newMod(d, "d3")

	a.NotError(d.Init(log.Default(), "default"))
	a.Equal(len(inits), 4).
		Equal(inits["m1"], 1). // 为 1 表示不会被多次调用
		Equal(inits["d1"], 1).
		Equal(inits["d2"], 1).
		Equal(inits["d3"], 1)

	a.NotError(d.Init(log.Default(), "default"))
}
