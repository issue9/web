// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"testing"

	"github.com/issue9/assert"
)

func TestConfig_init(t *testing.T) {
	a := assert.New(t)
	cfg := &Config{HTTPS: true}

	// 正常加载之后，测试各个变量是否和配置文件中的一样。
	a.NotPanic(func() { cfg.init() })
	a.Equal(":443", cfg.Port).
		Equal("", cfg.ServerName).
		True(cfg.HTTPS)
}
