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
	"github.com/issue9/mux"

	"github.com/issue9/web/internal/app/webconfig"
)

var muxtest = mux.New(false, false, nil, nil)

func TestNew(t *testing.T) {
	a := assert.New(t)

	ms, err := New(muxtest, &webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)
	a.Equal(len(ms.Modules()), 1).
		Equal(ms.modules[0].Name, coreModuleName)

	// 以下内容需要用到 plugin 功能，该功能 windows 并未实例，暂时跳过
	if !isPluginOS() {
		return
	}

	ms, err = New(muxtest, &webconfig.WebConfig{
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
	ms, err := New(muxtest, &webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)

	m1 := ms.NewModule("users1", "user1 module", "users2", "users3")
	m1.NewTag("v1").
		AddInitTitle("安装数据表 users1", func() error { return errors.New("falid message") })

	m2 := ms.NewModule("users2", "user2 module", "users3")
	m2.NewTag("v1").AddInitTitle("安装数据表 users2", func() error { return nil })

	m3 := ms.NewModule("users3", "user3 mdoule")
	tag := m3.NewTag("v1")
	tag.AddInitTitle("安装数据表 users3-1", func() error { return nil })
	tag.AddInitTitle("安装数据表 users3-2", func() error { return nil })
	tag.AddInitTitle("安装数据表 users3-3", func() error { return nil })

	a.NotError(ms.Init("install", log.New(os.Stderr, "", 0)))
	a.Error(ms.Init("v1", nil))
	a.NotError(ms.Init("not exists", nil))
}
