// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package modules

import (
	"errors"
	"log"
	"os"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/internal/webconfig"
)

func TestNew(t *testing.T) {
	a := assert.New(t)

	ms, err := New(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)
	a.Equal(len(ms.Modules()), 1).
		Equal(ms.modules[0].Name, CoreModuleName)

	a.NotNil(ms.Mux()).
		Equal(ms.Mux(), ms.router.Mux())

	// 以下内容需要用到 plugin 功能，该功能 windows 并未实例，暂时跳过
	if !isPluginOS() {
		return
	}

	ms, err = New(&webconfig.WebConfig{
		Plugins: "./testdata/plugin_*.so",
		Static: map[string]string{
			"/url": "/path",
		},
	})
	a.NotError(err).NotNil(ms)
	a.Equal(len(ms.Modules()), 3) // 插件*2 + 默认的模块
}

func TestModules_Init(t *testing.T) {
	a := assert.New(t)
	ms, err := New(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)

	m1 := ms.NewModule("users1", "user1 module", "users2", "users3")
	m1.NewTag("v1").
		AddInit(func() error { return errors.New("falid message") }, "安装数据表 users1")

	m2 := ms.NewModule("users2", "user2 module", "users3")
	m2.NewTag("v1").AddInit(func() error { return nil }, "安装数据表 users2")

	m3 := ms.NewModule("users3", "user3 mdoule")
	tag := m3.NewTag("v1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-2")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-3")

	a.NotError(ms.Init("install", log.New(os.Stderr, "", 0)))
	a.Error(ms.Init("v1", nil))
	a.NotError(ms.Init("not exists", nil))
}

func TestModules_Tags(t *testing.T) {
	a := assert.New(t)
	ms, err := New(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)

	m1 := ms.NewModule("users1", "user1 module", "users2", "users3")
	m1.NewTag("v1").
		AddInit(func() error { return errors.New("falid message") }, "安装数据表 users1")
	m1.NewTag("v2")

	m2 := ms.NewModule("users2", "user2 module", "users3")
	m2.NewTag("v1").AddInit(func() error { return nil }, "安装数据表 users2")
	m2.NewTag("v3")

	m3 := ms.NewModule("users3", "user3 mdoule")
	tag := m3.NewTag("v1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-2")
	m3.NewTag("v4")

	tags := ms.Tags()
	a.Equal(tags, []string{"v1", "v2", "v3", "v4"})
}
