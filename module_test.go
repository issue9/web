// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func clearModules() {
	modulesMu.Lock()
	defer modulesMu.Unlock()

	modules = map[string]*Module{}
}

func TestModules(t *testing.T) {
	a := assert.New(t)
	clearModules()

	m1, err := NewModule("m1")
	a.NotError(err).NotNil(m1)
	a.Equal(Modules(), modules).
		Equal(modules, map[string]*Module{"m1": m1})

	m2, err := NewModule("m2")
	a.NotError(err).NotNil(m2)
	a.Equal(Modules(), modules).
		Equal(modules, map[string]*Module{"m1": m1, "m2": m2})
}

func TestNewModule(t *testing.T) {
	a := assert.New(t)
	clearModules()

	m1, err := NewModule("m1")
	a.NotError(err).NotNil(m1)

	// 添加相同名称的Module
	m1, err = NewModule("m1")
	a.Nil(m1).
		Equal(err, ErrModuleExists).
		Equal(1, len(modules))

	// 依赖m1
	m2, err := NewModule("m2", "m1")
	a.NotNil(m2).NotError(err)

	// 依赖m1,m2
	m3, err := NewModule("m3", "m1", "m2")
	a.NotNil(m3).NotError(err)

	// 依赖项，并不完全存在
	m4, err := NewModule("m4", "m1", "m2", "notexists")
	a.Nil(m4).Error(err)

	// TODO 循环依赖测试
}

func TestGetModule(t *testing.T) {
	a := assert.New(t)
	clearModules()

	m1, err := NewModule("m1")
	a.NotError(err).NotNil(m1)

	m := GetModule("m1")
	a.Equal(m, m1)

	// 不存在的Module名称
	m = GetModule("M1")
	a.Nil(m)

	// 空的Module名称
	m = GetModule("")
	a.Nil(m)
}

func TestModule_Status(t *testing.T) {
	a := assert.New(t)
	clearModules()

	m1, err := NewModule("m1")
	a.NotError(err).NotNil(m1)

	m2, err := NewModule("m2")
	a.NotError(err).NotNil(m2)

	// 默认状态
	a.True(m1.IsRunning()).True(m2.IsRunning())

	m1.Stop()
	a.False(m1.IsRunning()).True(m2.IsRunning())

	m1.Stop()
	m2.Stop()
	a.False(m1.IsRunning()).False(m2.IsRunning())

	m2.Start()
	a.False(m1.IsRunning()).True(m2.IsRunning())
}

func TestModule_Add(t *testing.T) {
	a := assert.New(t)
	clearModules()

	fn := func(w http.ResponseWriter, req *http.Request) {}
	h := http.HandlerFunc(fn)

	m, err := NewModule("m")
	a.NotError(err).NotNil(m)

	a.NotPanic(func() { m.Get("h", h) })
	a.NotPanic(func() { m.Post("h", h) })
	a.NotPanic(func() { m.Put("h", h) })
	a.NotPanic(func() { m.Delete("h", h) })
	a.NotPanic(func() { m.Patch("h", h) })
	a.NotPanic(func() { m.Any("anyH", h) })
	a.NotPanic(func() { m.GetFunc("fn", fn) })
	a.NotPanic(func() { m.PostFunc("fn", fn) })
	a.NotPanic(func() { m.PutFunc("fn", fn) })
	a.NotPanic(func() { m.DeleteFunc("fn", fn) })
	a.NotPanic(func() { m.PatchFunc("fn", fn) })
	a.NotPanic(func() { m.AnyFunc("anyFN", fn) })

	// 添加相同的pattern
	a.Panic(func() { m.Any("h", h) })

	// handler不能为空
	a.Panic(func() { m.Add("abc", nil, "GET") })
	// pattern不能为空
	a.Panic(func() { m.Add("", h, "GET") })
	// 不支持的methods
	a.Panic(func() { m.Add("abc", h, "GET123") })
}
