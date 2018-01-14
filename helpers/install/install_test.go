// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package install

import (
	"errors"
	"testing"

	"github.com/issue9/assert"
)

func TestInstall(t *testing.T) {
	a := assert.New(t)

	i1 := New("users1", "users2", "users3")
	i1.Event("安装数据表users", func() *Return { return ReturnMessage("默认用户为admin:123") })

	i2 := New("users2", "users3")
	i2.Event("安装数据表users", func() *Return { return nil })

	i3 := New("users3")
	i3.Event("安装数据表users", func() *Return { return nil })
	i3.Event("安装数据表users", func() *Return { return ReturnError(errors.New("falid message")) })
	i3.Event("安装数据表users", func() *Return { return nil })

	Add(i1)
	Add(i2)
	Add(i3)
	a.NotError(Install())
}
