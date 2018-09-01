// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package modules

import (
	"errors"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux"

	"github.com/issue9/web/internal/app/webconfig"
)

var muxtest = mux.New(false, false, nil, nil)

func TestNew(t *testing.T) {
	a := assert.New(t)

	ms, err := New(10, muxtest, &webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)
	a.Equal(len(ms.Modules()), 1).
		Equal(ms.modules[0].Name, CoreModuleName)

	ms, err = New(10, muxtest, &webconfig.WebConfig{
		Plugins: "./testdata/plugin_*.so",
		Static: map[string]string{
			"/url": "/path",
		},
	})
	a.NotError(err).NotNil(ms)
	a.Equal(len(ms.Modules()), 3) // 插件*2 + 默认的模块
}

func TestModules_Install(t *testing.T) {
	a := assert.New(t)
	ms, err := New(10, muxtest, &webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)

	m1 := ms.NewModule("users1", "user1 module", "users2", "users3")
	m1.NewTag("v1").
		Task("安装数据表users", func() error { return errors.New("默认用户为admin:123") })

	m2 := ms.NewModule("users2", "user2 module", "users3")
	m2.NewTag("v1").Task("安装数据表users", func() error { return nil })

	m3 := ms.NewModule("users3", "user3 mdoule")
	tag := m3.NewTag("v1")
	tag.Task("安装数据表users", func() error { return nil })
	tag.Task("安装数据表users", func() error { return errors.New("falid message") })
	tag.Task("安装数据表users", func() error { return nil })

	a.NotError(ms.Install("install"))
	a.NotError(ms.Install("not exists"))
}
