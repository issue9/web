// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/version"
)

func TestVersion(t *testing.T) {
	assert.True(t, version.SemVerValid(Version), "无效的版本号")
}

func TestCheckVersion(t *testing.T) {
	a := assert.New(t)

	checkVersion("go1.10")
	checkVersion("go1.10.1")
	checkVersion("1.11.1")

	checkVersion("devel +12345")

	// 错误的语法
	a.Panic(func() {
		checkVersion("1.ab.")
	})

	// 版本过低
	a.Panic(func() {
		checkVersion("go1.9")
	})
}
