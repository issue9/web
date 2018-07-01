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

	m := New(router, "users2", "users2 mdoule")
	m.Task("v1", "安装数据表users", func() error { return nil })
	m.Task("v1", "安装数据表users", func() error { return errors.New("falid message") })
	m.Task("v1", "安装数据表users", func() error { return nil })

	f := m.GetInstall("v1")
	a.NotNil(f)
	a.NotError(f())
	f = m.GetInstall("not-exists")
	a.NotNil(f)
	a.NotError(f())
}
