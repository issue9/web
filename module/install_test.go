// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"testing"

	"github.com/issue9/assert"
)

func TestNewMessage(t *testing.T) {
	a := assert.New(t)

	err := NewMessage("test")
	a.Error(err)
	msg, ok := err.(message)
	a.True(ok).
		Equal(string(msg), "test")
}

func TestVersion(t *testing.T) {
	a := assert.New(t)
	m := New("user1", "user1 desc")
	a.NotNil(m)

	v := m.NewVersion("0.1.0")
	a.NotNil(v)
	v.Task("title1", nil)
	a.Equal(m.Installs["0.1.0"][0].title, "title1")
}

func TestModule_GetInstall(t *testing.T) {
	a := assert.New(t)

	m := New("users2", "users2 mdoule")
	a.NotNil(m)

	m.Task("v1", "安装数据表users", func() error { return NewMessage("success message") })
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
