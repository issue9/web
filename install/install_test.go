// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package install

import (
	"errors"
	"testing"
)

func TestInstall(t *testing.T) {
	i := New("users1", "users2", "users3")
	i.Event("安装数据表users", func() *Return { return ReturnMessage("默认用户为admin:123") })
	i.Done()

	i = New("users2", "users4")
	i.Event("安装数据表users", func() *Return { return nil })
	i.Done()

	// Events
	i = New("users3")
	i.Event("安装数据表users", func() *Return { return nil })
	i.Events(map[string]func() *Return{
		"安装数据表users1": func() *Return { return ReturnOK() },
		"安装数据表users2": func() *Return { return nil },
		"安装数据表users3": func() *Return { return nil },
	})
	i.Done()

	i = New("users4")
	i.Event("安装数据表users", func() *Return { return nil })
	i.Event("安装数据表users", func() *Return { return ReturnError(errors.New("falid message")) })
	i.Event("安装数据表users", func() *Return { return nil })
	i.Done()

	i = New("users5")
	i.Events(map[string]func() *Return{
		"安装数据表users1": func() *Return { return nil },
		"安装数据表users2": func() *Return { return ReturnError(errors.New("falid message")) },
		"安装数据表users3": func() *Return { return nil },
	})
	i.Done()

	Install()
}
