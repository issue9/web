// SPDX-License-Identifier: MIT

package dep

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"github.com/issue9/assert"
)

func TestDep_NewModule(t *testing.T) {
	a := assert.New(t)
	dep := New()
	a.NotNil(dep)

	m1, err := dep.NewModule("m1", "1.0.0", "m1 desc")
	a.NotError(err).NotNil(m1)
	m2, err := dep.NewModule("m2", "1.0.0", "m2 desc", "m1")
	a.NotError(err).NotNil(m2).
		Equal(m2.Version(), "1.0.0").
		Equal(m2.ID(), "m2")

	m11, err := dep.NewModule("m1", "1.0.0", "m1 desc")
	a.ErrorString(err, "存在同名的模块").Nil(m11)

	a.Equal(m1.ID(), "m1").Equal(m1.Description(), "m1 desc").Empty(m1.Deps())
	a.Equal(m2.ID(), "m2").Equal(m2.Description(), "m2 desc").Equal(m2.Deps(), []string{"m1"})

	ms := dep.Modules()
	a.Equal(len(ms), 2)
}

func TestModule_Init(t *testing.T) {
	a := assert.New(t)
	dep := New()
	a.NotNil(dep)

	m1, err := dep.NewModule("m1", "1.0.0", "m1 desc")
	a.NotError(err).NotNil(m1)

	buf := new(bytes.Buffer)
	t1 := m1.Tag("t1")
	t1.AddInit("1", func() error { return buf.WriteByte('1') }).
		AddInit("2", func() error { return buf.WriteByte('2') })
	a.NotError(m1.Init(log.Default(), "t1"))
	a.True(m1.Inited("t1")).Equal(buf.String(), "12")

	// 第二次不再调用
	buf.Reset()
	a.NotError(m1.Init(log.Default(), "t1"))
	a.True(m1.Inited("t1")).Equal(buf.String(), "")

	// 函数返回错误

	m2, err := dep.NewModule("m2", "1.0.0", "m2 desc")
	a.NotError(err).NotNil(m2)

	buf.Reset()
	t2 := m2.Tag("t2")
	t2.AddInit("1", func() error { return buf.WriteByte('1') }).
		AddInit("2", func() error { return errors.New("error at 2") }).
		AddInit("3", func() error { return buf.WriteByte('3') })

	a.ErrorString(m2.Init(log.Default(), "t2"), "error at 2")
	a.False(m2.Inited("t2")).
		Equal(buf.String(), "1")
}
