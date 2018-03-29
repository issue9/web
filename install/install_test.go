// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package install

import (
	"errors"
	"testing"

	"github.com/issue9/assert"
)

func TestModule_Get(t *testing.T) {
	a := assert.New(t)
	m := New("m1", "m2")

	i := m.Get("install")
	a.NotNil(i)
	a.Equal(len(m.tasks), 1)

	i = m.Get("install")
	a.NotNil(i)
	a.Equal(len(m.tasks), 1)

	i = m.Get("v1.0")
	a.NotNil(i)
	a.Equal(len(m.tasks), 2)

	// 清空
	modules = make([]*Module, 0, 10)
}

func TestInstall(t *testing.T) {
	a := assert.New(t)

	i1 := New("users1", "users2", "users3").Get("install")
	i1.Task("安装数据表users", func() *Return { return ReturnMessage("默认用户为admin:123") })

	i2 := New("users2", "users3").Get("install")
	i2.Task("安装数据表users", func() *Return { return nil })

	i3 := New("users3").Get("install")
	i3.Task("安装数据表users", func() *Return { return nil }).
		Task("安装数据表users", func() *Return { return ReturnError(errors.New("falid message")) }).
		Task("安装数据表users", func() *Return { return nil })

	a.NotError(Install("install"))
	a.NotError(Install("not exists"))
}
