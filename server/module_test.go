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
	var count int

	m1, err := s.NewModule("users1", "1.0.0", localeutil.Phrase("user1 module"), "users2", "users3")
	a.NotNil(m1).NotError(err)
	m1.AddInit("init m1", func() error { count++; return nil })   // count++
	m1.AddUninit("init m1", func() error { count--; return nil }) // count--
	v1 := m1.Action("v1")
	a.NotNil(v1)
	v1.AddInit("安装数据表 users1", func() error { return errors.New("failed message") })
	v1.AddUninit("安装数据表 users1", func() error { return nil })

	m2, err := s.NewModule("users2", "1.0.0", localeutil.Phrase("user2 module"), "users3")
	a.NotNil(m2).NotError(err)
	v2 := m2.Action("v2")
	a.NotNil(v2)
	v2.AddInit("安装数据表 users2", func() error { count++; return nil }) // count++
	m2.Action("v3")

	m3, err := s.NewModule("users3", "1.0.0", localeutil.Phrase("user3 module"))
	a.NotNil(m3).NotError(err)
	v2 = m3.Action("v2")
	a.NotNil(v2)
	v2.AddInit("安装数据表 users3-1", func() error { count++; return nil }) // count++
	v2.AddInit("安装数据表 users3-2", func() error { count++; return nil }) // count++
	a.NotNil(m3.Action("v4"))

	actions := s.Actions()
	a.Equal(actions, []string{"v1", "v2", "v3", "v4"})

	a.Panic(func() {
		s.initModules("", false) // 空值
	})

	a.Equal(0, count)
	a.ErrorString(s.initModules("v1", false), "failed message") // 这里执行了一次 m1.inits 中的 count++

	a.Equal(1, count)
	a.NotError(s.initModules("v1", true)) // 这里执行了一次 m1.uninits 中的 count--
	a.Equal(0, count)

	a.NotError(s.initModules("v2", false)) // 包括了 m1.inits 中的 count++
	a.Equal(4, count)
	a.NotError(s.initModules("not-exists", false))
}

func TestPluginInitFuncName(t *testing.T) {
	a := assert.New(t)

	a.True(unicode.IsUpper(rune(PluginInitFuncName[0])))
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

func TestAction_AddInit(t *testing.T) {
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

func TestModule_Object(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)
	o := 5

	m1, err := s.NewModule("m1", "1.0.0", localeutil.Phrase("m1 desc"))
	a.NotError(err).NotNil(m1)
	m1.AttachObject(o)
	a.Equal(m1.Object(), o)

	m2, err := s.NewModule("m2", "1.0.0", localeutil.Phrase("m2 desc"), "m1")
	a.NotError(err).NotNil(m2)
	a.Equal(m2.DepObject("m1"), o)
	a.PanicString(func() {
		m2.DepObject("not-dep")
	}, "的依赖对象")
	a.PanicString(func() {
		m1.DepObject("m1")
	}, "的依赖对象")

	// m2 并未指定 Object
	m3, err := s.NewModule("m3", "1.0.0", localeutil.Phrase("m3 desc"), "not-exists")
	a.NotError(err).NotNil(m3)
	a.PanicString(func() {
		m3.DepObject("not-exists")
	}, "依赖项 not-exists 未找到")
}
