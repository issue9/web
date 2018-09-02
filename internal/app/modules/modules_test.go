// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package modules

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux"

	"github.com/issue9/web/internal/app/webconfig"
	"github.com/issue9/web/module"
)

var (
	muxtest = mux.New(false, false, nil, nil)
	router  = muxtest.Prefix("")

	f1 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
)

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

func TestModules_Init(t *testing.T) {
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

	a.NotError(ms.Init("install"))
	a.NotError(ms.Init("not exists"))
}

func TestGetInit(t *testing.T) {
	a := assert.New(t)

	m := module.New("m1", "m1 desc")
	a.NotNil(m)
	fn := getInit(m, router, "")
	a.NotNil(fn).NotError(fn())

	// 返回错误
	m = module.New("m2", "m2 desc")
	a.NotNil(m)
	m.AddInit(func() error {
		return errors.New("error")
	})
	fn = getInit(m, router, "")
	a.NotNil(fn).ErrorString(fn(), "error")

	w := new(bytes.Buffer)
	m = module.New("m3", "m3 desc")
	a.NotNil(m)
	m.AddInit(func() error {
		_, err := w.WriteString("m3")
		return err
	})
	m.GetFunc("/get", f1)
	m.Prefix("/p").PostFunc("/post", f1)
	fn = getInit(m, router, "")
	a.NotNil(fn).
		NotError(fn()).
		Equal(w.String(), "m3")
}

func TestModule_GetInit2(t *testing.T) {
	a := assert.New(t)

	m := module.New("users2", "users2 mdoule")
	a.NotNil(m)

	tag := m.NewTag("v1")
	tag.Task("安装数据表users", func() error { return nil })
	tag.Task("安装数据表users", func() error { return nil })

	f := getInit(m, router, "v1")
	a.NotNil(f)
	a.NotError(f())
	f = getInit(m, router, "not-exists")
	a.NotNil(f)
	a.NotError(f())
}
