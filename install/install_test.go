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

	i := New("users1", "users2", "users3")
	i.Event("安装数据表users", func() *Return { return ReturnMessage("默认用户为admin:123") })
	i.Done()

	i = New("users2", "users3")
	i.Event("安装数据表users", func() *Return { return nil })
	i.Done()

	i = New("users3")
	i.Event("安装数据表users", func() *Return { return nil })
	i.Event("安装数据表users", func() *Return { return ReturnError(errors.New("falid message")) })
	i.Event("安装数据表users", func() *Return { return nil })
	i.Done()

	a.NotError(Install())
}
