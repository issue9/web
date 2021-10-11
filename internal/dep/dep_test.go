// SPDX-License-Identifier: MIT

package dep

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/localeutil"
)

func TestDep(t *testing.T) {
	a := assert.New(t)

	m1 := NewItem("users1", []string{"users2", "users3"}, []Executor{{Title: "init users1", F: func() error { return errors.New("failed message") }}})
	m2 := NewItem("users2", []string{"users3"}, []Executor{{Title: "init users2", F: func() error { return nil }}})
	m3 := NewItem("users3", nil, []Executor{{Title: "init users3", F: func() error { return nil }}})

	a.NotError(Dep(log.Default(), []*Item{m2, m3}))
	a.ErrorString(Dep(log.Default(), []*Item{m2, m3, m1}), "failed message")
}

func TestIsDep(t *testing.T) {
	a := assert.New(t)

	m1 := NewItem("m1", []string{"d1", "d2"}, nil)
	d1 := NewItem("d1", []string{"d3"}, nil)
	items := []*Item{m1, d1}

	a.True(IsDep(items, "m1", "d1"))
	a.True(IsDep(items, "m1", "d2"))
	a.True(IsDep(items, "m1", "d3")) // 通过 d1 继承
	a.False(IsDep(items, "m1", "m1"))

	// 循环依赖
	m1 = &Item{id: "m1", deps: []string{"d1", "d2"}}
	d1 = &Item{id: "d1", deps: []string{"d3"}}
	d3 := &Item{id: "d3", deps: []string{"d1"}}
	items = []*Item{m1, d1, d3}
	a.True(IsDep(items, "d1", "d1"))

	// 不存在的模块
	a.False(IsDep(items, "d10", "d1"))
}

func TestCheckDeps(t *testing.T) {
	a := assert.New(t)

	m1 := NewItem("m1", []string{"d1", "d2"}, nil)
	d1 := NewItem("d1", []string{"d3"}, nil)
	items := []*Item{m1, d1}

	mm := findItem(items, "m1")
	a.NotNil(mm).Equal(mm, m1).
		Equal(checkDeps(items, m1), localeutil.Error("not found dependence %s", "m1", "d2")) // 依赖项不存在

	m1 = NewItem("m1", []string{"d1", "d2"}, nil)
	d1 = NewItem("d1", []string{"d3"}, nil)
	d2 := NewItem("d2", []string{"d3"}, nil)
	items = []*Item{m1, d1, d2}
	a.NotError(checkDeps(items, m1))

	// 循环依赖
	m1 = NewItem("m1", []string{"d1", "d2"}, nil)
	d1 = NewItem("d1", []string{"d3"}, nil)
	d2 = NewItem("d2", []string{"d3"}, nil)
	d3 := NewItem("d3", []string{"d2"}, nil)
	items = []*Item{m1, d1, d2, d3}
	a.NotNil(findItem(items, "d2")).
		Equal(checkDeps(items, d2), localeutil.Error("%s cyclic dependence", "d2"))
}

func TestInitItem(t *testing.T) {
	a := assert.New(t)
	buf := new(bytes.Buffer)

	m1 := NewItem("m1", nil, []Executor{
		{
			Title: "1",
			F:     func() error { return buf.WriteByte('1') },
		}, {
			Title: "2",
			F:     func() error { return buf.WriteByte('2') },
		},
	})

	a.NotError(initItem([]*Item{m1}, m1, log.Default()))
	a.True(m1.called).Equal(buf.String(), "12")

	// 第二次不再调用
	buf.Reset()
	a.NotError(initItem([]*Item{m1}, m1, log.Default()))
	a.True(m1.called).Equal(buf.String(), "")

	// 函数返回错误

	buf.Reset()
	m1 = &Item{id: "m2"}
	m1 = NewItem("m2", nil, []Executor{
		{
			Title: "1",
			F:     func() error { return buf.WriteByte('1') },
		}, {
			Title: "2",
			F:     func() error { return errors.New("error at 2") },
		}, {
			Title: "3",
			F:     func() error { return buf.WriteByte('3') },
		},
	})

	a.ErrorString(initItem([]*Item{m1}, m1, log.Default()), "error at 2")
	a.False(m1.called).
		Equal(buf.String(), "1")
}
