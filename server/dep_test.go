// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"github.com/issue9/assert"
)

func TestServer_InitModules(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)

	m1, err := s.NewModule("users1", "1.0.0", "user1 module", "users2", "users3")
	a.NotNil(m1).NotError(err)
	t1 := m1.Tag("v1")
	a.NotNil(t1)
	t1.AddInit("安装数据表 users1", func() error { return errors.New("failed message") })
	m1.Tag("v2")

	m2, err := s.NewModule("users2", "1.0.0", "user2 module", "users3")
	a.NotNil(m2).NotError(err)
	t2 := m2.Tag("v1")
	a.NotNil(t2)
	t2.AddInit("安装数据表 users2", func() error { return nil })
	m2.Tag("v3")

	m3, err := s.NewModule("users3", "1.0.0", "user3 module")
	a.NotNil(m3).NotError(err)
	tag := m3.Tag("v1")
	a.NotNil(tag)
	tag.AddInit("安装数据表 users3-1", func() error { return nil })
	tag.AddInit("安装数据表 users3-2", func() error { return nil })
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

func TestServer_Tags(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	a.Empty(srv.Tags())

	m1, err := srv.NewModule("m1", "1.0.0", "m1 desc")
	a.NotError(err).NotNil(m1)
	a.NotNil(m1.Tag("d1"))
	a.NotNil(m1.Tag("d2"))
	a.NotNil(m1.Tag("d2"))

	m2, err := srv.NewModule("m2", "1.0.0", "m2 desc")
	a.NotError(err).NotNil(m2)
	a.NotNil(m2.Tag("d2"))
	a.NotNil(m2.Tag("d0"))

	m3, err := srv.NewModule("m3", "1.0.0", "m3 desc")
	a.NotError(err).NotNil(m3)

	a.Equal(srv.Tags(), []string{"d0", "d1", "d2"}).
		Equal(srv.Modules(), []*Module{m1, m2, m3})
}

func TestServer_isDep(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	a.Empty(srv.Tags())

	m, err := srv.NewModule("m1", "1.0.0", "m1 desc", "d1", "d2")
	a.NotError(err).NotNil(m)
	m, err = srv.NewModule("d1", "1.0.0", "d1 desc", "d3")
	a.NotError(err).NotNil(m)

	a.True(srv.isDep("m1", "d1"))
	a.True(srv.isDep("m1", "d2"))
	a.True(srv.isDep("m1", "d3")) // 通过 d1 继承
	a.False(srv.isDep("m1", "m1"))

	// 循环依赖
	srv = newServer(a)
	a.Empty(srv.Tags())
	m, err = srv.NewModule("m1", "1.0.0", "m1 desc", "d1", "d2")
	a.NotError(err).NotNil(m)
	m, err = srv.NewModule("d1", "1.0.0", "d1 desc", "d3")
	a.NotError(err).NotNil(m)
	m, err = srv.NewModule("d3", "1.0.0", "d3 desc", "d1")
	a.NotError(err).NotNil(m)
	a.True(srv.isDep("d1", "d1"))

	// 不存在的模块
	a.False(srv.isDep("d10", "d1"))
}

func TestServer_checkDeps(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	a.Empty(srv.Tags())

	srv.NewModule("m1", "1.0.0", "m1 desc", "d1", "d2")
	srv.NewModule("d1", "1.0.0", "d1 desc", "d3")

	m1 := srv.findModule("m1")
	a.NotNil(m1).
		ErrorString(srv.checkDeps(m1), "未找到") // 依赖项不存在

	srv = newServer(a)
	a.NotNil(srv)
	srv.NewModule("m1", "1.0.0", "m1 desc", "d1", "d2")
	srv.NewModule("d1", "1.0.0", "d1 desc", "d3")
	srv.NewModule("d2", "1.0.0", "d2 desc", "d3")
	a.NotError(srv.checkDeps(m1))

	// 自我依赖
	srv = newServer(a)
	a.NotNil(srv)
	srv.NewModule("m1", "1.0.0", "m1 desc", "d1", "d2")
	srv.NewModule("d1", "1.0.0", "d1 desc", "d3")
	srv.NewModule("d2", "1.0.0", "d2 desc", "d3")
	srv.NewModule("d3", "1.0.0", "d3 desc", "d2")
	d2 := srv.findModule("d2")
	a.NotNil(d2).
		ErrorString(srv.checkDeps(d2), "循环依赖自身")
}

func TestServer_initModule(t *testing.T) {
	a := assert.New(t)
	dep := newServer(a)
	a.NotNil(dep)

	m1, err := dep.NewModule("m1", "1.0.0", "m1 desc")
	a.NotError(err).NotNil(m1)

	buf := new(bytes.Buffer)
	t1 := m1.Tag("t1")
	t1.AddInit("1", func() error { return buf.WriteByte('1') }).
		AddInit("2", func() error { return buf.WriteByte('2') })
	a.NotError(dep.initModule(m1, log.Default(), "t1"))
	a.True(m1.Inited("t1")).Equal(buf.String(), "12")

	// 第二次不再调用
	buf.Reset()
	a.NotError(dep.initModule(m1, log.Default(), "t1"))
	a.True(m1.Inited("t1")).Equal(buf.String(), "")

	// 函数返回错误

	m2, err := dep.NewModule("m2", "1.0.0", "m2 desc")
	a.NotError(err).NotNil(m2)

	buf.Reset()
	t2 := m2.Tag("t2")
	t2.AddInit("1", func() error { return buf.WriteByte('1') }).
		AddInit("2", func() error { return errors.New("error at 2") }).
		AddInit("3", func() error { return buf.WriteByte('3') })

	a.ErrorString(dep.initModule(m2, log.Default(), "t2"), "error at 2")
	a.False(m2.Inited("t2")).
		Equal(buf.String(), "1")
}
