// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"io/fs"
	"log"
	"testing"
	"unicode"

	"github.com/issue9/assert"
)

var _ fs.FS = &Module{}

func TestPluginInitFuncName(t *testing.T) {
	a := assert.New(t)

	a.True(unicode.IsUpper(rune(PluginInitFuncName[0])))
}

func TestNewModule(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)

	m, err := s.NewModule("id", "1.0.0", "desc", "id1", "id2")
	a.NotNil(m).NotError(err).
		Equal(m.ID(), "id").
		Equal(m.Description(), "desc").
		Equal(m.Deps(), []string{"id1", "id2"})
}

func TestModule_Tag(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)
	m, err := s.NewModule("user1", "1.0.0", "user1 desc")
	a.NotNil(m).NotError(err)

	v := m.Tag("0.1.0")
	a.NotNil(v)
	a.NotNil(v.AddInit("title1", func() error { return nil }))
	a.Equal(v.Name(), "0.1.0")

	vv := m.Tag("0.1.0")
	a.Equal(vv, v)

	v2 := m.Tag("0.2.0")
	a.NotEqual(v2, v)
}

func TestServer_NewModule(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	a.NotNil(srv)

	m1, err := srv.NewModule("m1", "1.0.0", "m1 desc")
	a.NotError(err).NotNil(m1)
	m2, err := srv.NewModule("m2", "1.0.0", "m2 desc", "m1")
	a.NotError(err).NotNil(m2).
		Equal(m2.Version(), "1.0.0").
		Equal(m2.ID(), "m2")

	m11, err := srv.NewModule("m1", "1.0.0", "m1 desc")
	a.ErrorString(err, "存在同名的模块").Nil(m11)

	a.Equal(m1.ID(), "m1").Equal(m1.Description(), "m1 desc").Empty(m1.Deps())
	a.Equal(m2.ID(), "m2").Equal(m2.Description(), "m2 desc").Equal(m2.Deps(), []string{"m1"})

	ms := srv.Modules()
	a.Equal(len(ms), 2)
}

func TestTag_AddInit(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)
	m, err := s.NewModule("m1", "1.0.0", "m1 desc")
	a.NotError(err).NotNil(m)

	tag := m.Tag("t1")
	tag.AddInit("1", func() error { return nil }).
		AddInit("2", func() error { return nil }).
		AddInit("2", func() error { return nil })
	a.Equal(3, len(tag.executors))
	a.Equal(tag.executors[0].title, "1")
	a.Equal(tag.executors[2].title, "2")
	a.Equal(tag.Module(), m)
	a.Equal(tag.Server(), s)
}

func TestTag_init(t *testing.T) {
	a := assert.New(t)

	tag := &Tag{executors: make([]executor, 0, 5)}
	tag.AddInit("1", func() error { return nil }).
		AddInit("2", func() error { return nil })

	a.False(tag.Inited())
	a.NotError(tag.init(log.Default()))
	a.True(tag.Inited())
	a.NotError(tag.Inited())

	tag.AddInit("3", func() error { return nil })
	a.NotError(tag.init(log.Default()))
	a.True(tag.Inited())
	a.Equal(3, len(tag.executors))

	// failed

	tag = &Tag{executors: make([]executor, 0, 5)}
	tag.AddInit("1", func() error { return nil }).
		AddInit("2", func() error { return errors.New("error at 2") })
	a.False(tag.Inited())
	a.ErrorString(tag.init(log.Default()), "error at 2")
}
