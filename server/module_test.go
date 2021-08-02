// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"testing"
	"unicode"

	"github.com/issue9/assert"
)

func TestModuleFuncName(t *testing.T) {
	a := assert.New(t)

	a.True(unicode.IsUpper(rune(ModuleFuncName[0])))
}

func TestNewModule(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)

	m, err := s.NewModule("id", "desc", "id1", "id2")
	a.NotNil(m).NotError(err).
		Equal(m.ID(), "id").
		Equal(m.Description(), "desc").
		Equal(m.Deps(), []string{"id1", "id2"})
}

func TestModule_Tag(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)
	m, err := s.NewModule("user1", "user1 desc")
	a.NotNil(m).NotError(err)

	v := m.Tag("0.1.0")
	a.NotNil(v)
	a.NotNil(v.On("title1", func() error { return nil }))

	vv := m.Tag("0.1.0")
	a.Equal(vv, v)

	v2 := m.Tag("0.2.0")
	a.NotEqual(v2, v)
}

func TestServer_InitModules(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)

	m1, err := s.NewModule("users1", "user1 module", "users2", "users3")
	a.NotNil(m1).NotError(err)
	t1 := m1.Tag("v1")
	a.NotNil(t1)
	t1.On("安装数据表 users1", func() error { return errors.New("failed message") })
	m1.Tag("v2")

	m2, err := s.NewModule("users2", "user2 module", "users3")
	a.NotNil(m2).NotError(err)
	t2 := m2.Tag("v1")
	a.NotNil(t2)
	t2.On("安装数据表 users2", func() error { return nil })
	m2.Tag("v3")

	m3, err := s.NewModule("users3", "user3 module")
	a.NotNil(m3).NotError(err)
	tag := m3.Tag("v1")
	a.NotNil(tag)
	tag.On("安装数据表 users3-1", func() error { return nil })
	tag.On("安装数据表 users3-2", func() error { return nil })
	a.NotNil(m3.Tag("v4"))

	tags := s.Tags()
	a.Equal(tags, []string{"v1", "v2", "v3", "v4"})

	a.Panic(func() {
		s.InitModules("") // 空值
	})
	a.ErrorString(s.InitModules("v1"), "failed message")
	a.NotError(s.InitModules("v2"))
	a.NotError(s.InitModules("not-exists"))
}
