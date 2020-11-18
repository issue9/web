// SPDX-License-Identifier: MIT

package dep

import (
	"log"
	"os"
	"testing"

	"github.com/issue9/assert"
)

func TestIsDep(t *testing.T) {
	a := assert.New(t)

	ms := []Module{
		newMod("m1", nil, "d1", "d2"),
		newMod("d1", nil, "d3"),
	}

	a.True(isDep(ms, "m1", "d1"))
	a.True(isDep(ms, "m1", "d2"))
	a.True(isDep(ms, "m1", "d3")) // 通过 d1 继承
	a.False(isDep(ms, "m1", "m1"))

	// 循环依赖
	ms = []Module{
		newMod("m1", nil, "d1", "d2"),
		newMod("d1", nil, "d3"),
		newMod("d3", nil, "d1"),
	}
	a.True(isDep(ms, "d1", "d1"))

	// 不存在的模块
	a.False(isDep(ms, "d10", "d1"))
}

func TestCheckDeps(t *testing.T) {
	a := assert.New(t)

	ms := []Module{
		newMod("m1", nil, "d1", "d2"),
		newMod("d1", nil, "d3"),
	}

	m1 := findModule(ms, "m1")
	a.NotNil(m1).
		Error(checkDeps(ms, m1)) // 依赖项不存在

	ms = []Module{
		newMod("m1", nil, "d1", "d2"),
		newMod("d1", nil, "d3"),
		newMod("d2", nil, "d3"),
	}
	a.NotError(checkDeps(ms, m1))

	// 自我依赖
	ms = []Module{
		newMod("m1", nil, "d1", "d2"),
		newMod("d1", nil, "d3"),
		newMod("d2", nil, "d3"),
		newMod("d3", nil, "d2"),
	}
	d2 := findModule(ms, "d2")
	a.NotNil(d2).
		Error(checkDeps(ms, d2))
}

func TestInit(t *testing.T) {
	a := assert.New(t)

	inits := map[string]int{}
	infolog := log.New(os.Stderr, "", 0)

	f := func(name string) func() {
		return func() {
			inits[name] = inits[name] + 1
		}
	}

	// 缺少依赖项 d3
	ms := []Module{
		newMod("m1", f("m1"), "d1", "d2"),
		newMod("d1", f("d1"), "d3"),
		newMod("d2", f("d2"), "d3"),
	}
	a.Error(Init(ms, infolog))

	ms = []Module{
		newMod("m1", f("m1"), "d1", "d2"),
		newMod("d1", f("d1"), "d3"),
		newMod("d2", f("d2"), "d3"),
		newMod("d3", f("d3")),
	}

	a.NotError(Init(ms, infolog))
	a.Equal(len(inits), 4).
		Equal(inits["m1"], 1).
		Equal(inits["d1"], 1).
		Equal(inits["d2"], 1)
}
