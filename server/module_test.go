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
	s := newServer(a, nil)
	var count int

	m1 := s.NewModule("users1", "1.0.0", localeutil.Phrase("user1 module"), "users2", "users3")
	a.NotNil(m1)
	m1.AddInit("init m1", func() error { count++; return nil })   // count++
	m1.AddUninit("init m1", func() error { count--; return nil }) // count--
	v1 := m1.Action("v1")
	a.NotNil(v1)
	v1.AddInit("安装数据表 users1", func() error { return errors.New("failed message") })
	v1.AddUninit("安装数据表 users1", func() error { return nil })

	m2 := s.NewModule("users2", "1.0.0", localeutil.Phrase("user2 module"), "users3")
	a.NotNil(m2)
	v2 := m2.Action("v2")
	a.NotNil(v2)
	v2.AddInit("安装数据表 users2", func() error { count++; return nil }) // count++
	m2.Action("v3")

	m3 := s.NewModule("users3", "1.0.0", localeutil.Phrase("user3 module"))
	a.NotNil(m3)
	v2 = m3.Action("v2")
	a.NotNil(v2)
	v2.AddInit("安装数据表 users3-1", func() error { count++; return nil }) // count++
	v2.AddInit("安装数据表 users3-2", func() error { count++; return nil }) // count++
	a.NotNil(m3.Action("v4"))

	actions := s.Actions()
	a.Equal(actions, []string{"v1", "v2", "v3", "v4"})

	a.Panic(func() {
		s.initModules(false, "") // 空值
	})

	a.Equal(0, count)
	a.ErrorString(s.initModules(false, "v1"), "failed message") // 这里执行了一次 m1.inits 中的 count++

	a.Equal(1, count)
	a.NotError(s.initModules(true, "v1")) // 这里执行了一次 m1.uninits 中的 count--
	a.Equal(0, count)

	a.NotError(s.initModules(false, "v2")) // 包括了 m1.inits 中的 count++
	a.Equal(4, count)
	a.NotError(s.initModules(false, "not-exists"))
}

func TestPluginInitFuncName(t *testing.T) {
	a := assert.New(t)

	a.True(unicode.IsUpper(rune(PluginInitFuncName[0])))
}

func TestServer_Actions(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a, nil)
	a.Empty(srv.Actions())

	m1 := srv.NewModule("m1", "1.0.0", localeutil.Phrase("m1 desc"))
	a.NotNil(m1)
	a.NotNil(m1.Action("d1"))
	a.NotNil(m1.Action("d2"))
	a.NotNil(m1.Action("d2"))

	m2 := srv.NewModule("m2", "1.0.0", localeutil.Phrase("m2 desc"))
	a.NotNil(m2)
	a.NotNil(m2.Action("d2"))
	a.NotNil(m2.Action("d0"))

	m3 := srv.NewModule("m3", "1.0.0", localeutil.Phrase("m3 desc"))
	a.NotNil(m3)

	a.Equal(srv.Actions(), []string{"d0", "d1", "d2"}).
		Equal(len(srv.Modules(message.NewPrinter(language.Und))), 3)
}

func TestModule_Action(t *testing.T) {
	a := assert.New(t)
	s := newServer(a, nil)
	m := s.NewModule("user1", "1.0.0", localeutil.Phrase("user1 desc"))
	a.NotNil(m)

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
	srv := newServer(a, nil)

	builder := catalog.NewBuilder()
	a.NotError(builder.SetString(language.SimplifiedChinese, "m1 desc", "m1 描述信息"))
	printer := message.NewPrinter(language.SimplifiedChinese, message.Catalog(builder))

	m1 := srv.NewModule("m1", "1.0.0", localeutil.Phrase("m1 desc"))
	a.NotNil(m1)
	m2 := srv.NewModule("m2", "1.0.0", localeutil.Phrase("m2 desc"), "m1")
	a.NotNil(m2)

	a.PanicString(func() {
		m := srv.NewModule("m1", "1.0.0", localeutil.Phrase("m1 desc"))
		a.Nil(m)
	}, "存在同名的模块")

	ms := srv.Modules(printer)
	a.Equal(ms, []*ModuleInfo{
		{ID: "m1", Version: "1.0.0", Description: "m1 描述信息"},
		{ID: "m2", Version: "1.0.0", Description: "m2 desc", Deps: []string{"m1"}},
	})
}

func TestAction_AddInit(t *testing.T) {
	a := assert.New(t)
	s := newServer(a, nil)
	m := s.NewModule("m1", "1.0.0", localeutil.Phrase("m1 desc"))
	a.NotNil(m)

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
	s := newServer(a, nil)
	o := 5

	m1 := s.NewModule("m1", "1.0.0", localeutil.Phrase("m1 desc"))
	a.NotNil(m1)
	m1.AttachObject(o)
	a.Equal(m1.object, o)

	m2 := s.NewModule("m2", "1.0.0", localeutil.Phrase("m2 desc"), "m1")
	a.NotNil(m2)
	a.Equal(m2.DepObject("m1"), o)

	// 间接依赖
	m3 := s.NewModule("m3", "1.0.0", localeutil.Phrase("m3 desc"), "m2")
	a.NotNil(m3)
	a.Equal(m3.DepObject("m1"), o)

	// 找不到依赖对象
	a.PanicString(func() {
		m2.DepObject("not-dep")
	}, "not-dep 并不是 m2 的依赖对象")

	// 循环依赖
	a.PanicString(func() {
		m1.DepObject("m1")
	}, "m1 并不是 m1 的依赖对象")

	// 不存在的依赖项
	m4 := s.NewModule("m4", "1.0.0", localeutil.Phrase("m4 desc"), "not-exists")
	a.NotNil(m4)
	a.PanicString(func() {
		m4.DepObject("not-exists")
	}, "依赖项 not-exists 未找到")
}
