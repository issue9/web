// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"io/fs"
	"log"
	"testing"
	"unicode"

	"github.com/issue9/assert"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

var _ fs.FS = &Module{}

func TestPluginInitFuncName(t *testing.T) {
	a := assert.New(t)

	a.True(unicode.IsUpper(rune(PluginInitFuncName[0])))
}

func TestNewModule(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)

	m, err := s.NewModule("id", "1.0.0", localeutil.Phrase("desc"), "id1", "id2")
	a.NotNil(m).NotError(err)
}

func TestModule_Action(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)
	m, err := s.NewModule("user1", "1.0.0", localeutil.Phrase("user1 desc"))
	a.NotNil(m).NotError(err)

	v := m.Action("0.1.0")
	a.NotNil(v)
	a.NotNil(v.AddInit("title1", func() error { return nil }))
	a.Equal(v.Name(), "0.1.0")

	vv := m.Action("0.1.0")
	a.Equal(vv, v)

	v2 := m.Action("0.2.0")
	a.NotEqual(v2, v)
}

func TestServer_NewModule(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	a.NotNil(srv)

	builder := catalog.NewBuilder()
	a.NotError(builder.SetString(language.SimplifiedChinese, "m1 desc", "m1 描述信息"))
	printer := message.NewPrinter(language.SimplifiedChinese, message.Catalog(builder))

	m1, err := srv.NewModule("m1", "1.0.0", localeutil.Phrase("m1 desc"))
	a.NotError(err).NotNil(m1)
	m2, err := srv.NewModule("m2", "1.0.0", localeutil.Phrase("m2 desc"), "m1")
	a.NotError(err).NotNil(m2)

	m11, err := srv.NewModule("m1", "1.0.0", localeutil.Phrase("m1 desc"))
	a.ErrorString(err, "存在同名的模块").Nil(m11)

	ms := srv.Modules(printer)
	a.Equal(ms, []*ModuleInfo{
		{ID: "m1", Version: "1.0.0", Description: "m1 描述信息"},
		{ID: "m2", Version: "1.0.0", Description: "m2 desc", Deps: []string{"m1"}},
	})
}

func TestActioon_AddInit(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)
	m, err := s.NewModule("m1", "1.0.0", localeutil.Phrase("m1 desc"))
	a.NotError(err).NotNil(m)

	tag := m.Action("t1")
	tag.AddInit("1", func() error { return nil }).
		AddInit("2", func() error { return nil }).
		AddInit("2", func() error { return nil })
	a.Equal(3, len(tag.executors))
	a.Equal(tag.executors[0].title, "1")
	a.Equal(tag.executors[2].title, "2")
	a.Equal(tag.Module(), m)
	a.Equal(tag.Server(), s)
}

func TestAction_init(t *testing.T) {
	a := assert.New(t)

	tag := &Action{executors: make([]executor, 0, 5)}
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

	tag = &Action{executors: make([]executor, 0, 5)}
	tag.AddInit("1", func() error { return nil }).
		AddInit("2", func() error { return errors.New("error at 2") })
	a.False(tag.Inited())
	a.ErrorString(tag.init(log.Default()), "error at 2")
}
