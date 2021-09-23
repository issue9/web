// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"io/fs"
	"testing"
	"unicode"

	"github.com/issue9/assert"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

var _ fs.FS = &Module{}

func TestServer_initModules(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)

	m1, err := s.NewModule("users1", "1.0.0", localeutil.Phrase("user1 module"), "users2", "users3")
	a.NotNil(m1).NotError(err)
	t1 := m1.Action("v1")
	a.NotNil(t1)
	t1.AddInit("安装数据表 users1", func() error { return errors.New("failed message") })

	m2, err := s.NewModule("users2", "1.0.0", localeutil.Phrase("user2 module"), "users3")
	a.NotNil(m2).NotError(err)
	t2 := m2.Action("v2")
	a.NotNil(t2)
	t2.AddInit("安装数据表 users2", func() error { return nil })
	m2.Action("v3")

	m3, err := s.NewModule("users3", "1.0.0", localeutil.Phrase("user3 module"))
	a.NotNil(m3).NotError(err)
	action := m3.Action("v2")
	a.NotNil(action)
	action.AddInit("安装数据表 users3-1", func() error { return nil })
	action.AddInit("安装数据表 users3-2", func() error { return nil })
	a.NotNil(m3.Action("v4"))

	actions := s.Actions()
	a.Equal(actions, []string{"v1", "v2", "v3", "v4"})

	a.Panic(func() {
		s.initModules("") // 空值
	})
	a.ErrorString(s.initModules("v1"), "failed message")
	a.NotError(s.initModules("v2"))
	a.NotError(s.initModules("not-exists"))
}

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

func TestServer_Actions(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	a.Empty(srv.Actions())

	m1, err := srv.NewModule("m1", "1.0.0", localeutil.Phrase("m1 desc"))
	a.NotError(err).NotNil(m1)
	a.NotNil(m1.Action("d1"))
	a.NotNil(m1.Action("d2"))
	a.NotNil(m1.Action("d2"))

	m2, err := srv.NewModule("m2", "1.0.0", localeutil.Phrase("m2 desc"))
	a.NotError(err).NotNil(m2)
	a.NotNil(m2.Action("d2"))
	a.NotNil(m2.Action("d0"))

	m3, err := srv.NewModule("m3", "1.0.0", localeutil.Phrase("m3 desc"))
	a.NotError(err).NotNil(m3)

	a.Equal(srv.Actions(), []string{"d0", "d1", "d2"}).
		Equal(len(srv.Modules(message.NewPrinter(language.Und))), 3)
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
	a.Equal(3, len(tag.inits))
	a.Equal(tag.inits[0].Title, "1")
	a.Equal(tag.inits[2].Title, "2")
	a.Equal(tag.Module(), m)
	a.Equal(tag.Server(), s)
}
