// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package help

import (
	"testing"

	"github.com/issue9/assert"
)

func TestRegister(t *testing.T) {
	a := assert.New(t)
	a.Equal(len(usages), 1) // init 注册的 help
	a.Equal(usages["help"], usage)

	Register("x1", usage)
	a.Equal(len(usages), 2).
		Equal(usages["x1"], usage)

	// 重复添加
	a.Panic(func() { Register("x1", usage) })
	a.Equal(len(usages), 2).
		Equal(usages["x1"], usage)
}
