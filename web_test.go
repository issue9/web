// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"os"
	"testing"

	"github.com/issue9/assert"
)

var initErr error

func TestMain(m *testing.M) {
	initErr = Init("./testdata", nil)

	os.Exit(m.Run())
}

// 检测在 TestMain() 中的功能是否存在错误。
func TestInit(t *testing.T) {
	a := assert.New(t)

	a.NotError(initErr).NotNil(defaultApp)
}

func TestURL(t *testing.T) {
	a := assert.New(t)

	a.Equal(URL("test"), "https://caixw.io/test")
	a.Equal(URL("/test/file.jpg"), "https://caixw.io/test/file.jpg")
}
