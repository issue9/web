// SPDX-License-Identifier: MIT

package module

import (
	"bytes"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v2"
)

func TestNewModule(t *testing.T) {
	a := assert.New(t)

	m := NewModule("m", "m desc", "m1", "m2")
	a.NotNil(m)
	a.Equal(m.ID(), "m").
		Equal(m.Description(), "m desc").
		False(m.Inited())

	// deps 为一个副本，修改  m.deps 不影响已经返回的 m.Deps() 的值
	deps := m.Deps()
	a.Equal(deps, m.deps)
	m.deps = []string{"m3"}
	a.NotEqual(deps, m.deps)
}

func TestModule_Init(t *testing.T) {
	a := assert.New(t)

	m := NewModule("m", "m desc", "m1", "m2")
	a.NotNil(m)

	b1 := &bytes.Buffer{}
	m.AddInit("f1", func() error { return b1.WriteByte('1') }).
		AddInit("f2", func() error { return b1.WriteByte('2') })
	a.NotError(m.Init("t1", logs.New())) // t1 不存在
	a.Empty(b1.Bytes()).
		False(m.Inited())

	a.NotError(m.Init("", logs.New()))
	a.Equal(b1.String(), "12").
		True(m.Inited())
	a.ErrorString(m.Init("", logs.New()), "已经初始化") // 已经初始化

	// tags

	b1.Reset()
	m.GetTag("t1").AddInit("f3", func() error { return b1.WriteByte('3') }).
		AddInit("f4", func() error { return b1.WriteByte('4') })
	a.NotError(m.Init("t1", logs.New()))
	a.Equal(b1.String(), "34")
}

func TestModule_GetTag(t *testing.T) {
	a := assert.New(t)

	m := NewModule("m", "m desc", "m1", "m2")
	a.NotNil(m)
	t1 := m.GetTag("t1")
	a.NotNil(t1)
	a.Equal(t1, m.GetTag("t1"))
}
