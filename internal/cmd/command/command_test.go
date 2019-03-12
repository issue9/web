// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package command

import (
	"testing"

	"github.com/issue9/assert"
)

func TestRegister(t *testing.T) {
	a := assert.New(t)
	a.Equal(len(commands), 1) // init 注册的 help
	a.Equal(commands["help"].usage, helpUsage).
		Equal(commands["help"].do, helpDo)

	Register("x1", helpDo, helpUsage)
	a.Equal(len(commands), 2).
		Equal(commands["x1"].usage, helpUsage)

	// 重复添加
	a.Panic(func() { Register("x1", helpDo, helpUsage) })
	a.Equal(len(commands), 2).
		Equal(commands["x1"].usage, helpUsage)
}
