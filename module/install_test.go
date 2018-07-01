// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"testing"

	"github.com/issue9/assert"
)

func TestModule_GetInstall(t *testing.T) {
	a := assert.New(t)

	m1 := New(router, "users1", "users1 module", "users2")
	m1.Task("v1", "安装数据表users", func() error { return errors.New("默认用户为admin:123") })
	f := m1.GetInstall("v1")
	a.NotNil(f)
	a.Nil(m1.GetInstall("not-exists"))

	m3 := New(router, "users2", "users2 mdoule")
	m3.Task("v1", "安装数据表users", func() error { return nil })
	m3.Task("v1", "安装数据表users", func() error { return errors.New("falid message") })
	m3.Task("v1", "安装数据表users", func() error { return nil })

	a.NotNil(m1.GetInstall("v1"))
	a.Nil(m1.GetInstall("not-exists"))
}
