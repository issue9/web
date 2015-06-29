// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"path/filepath"
	"testing"

	"github.com/issue9/assert"
)

func TestInitConfigDir(t *testing.T) {
	a := assert.New(t)

	// 默认为空值
	a.Equal(len(configDir), 0)

	// 触发panic时，不会改变configDir的值
	a.Panic(func() { initConfigDir("./emptyDir") })
	a.Equal(len(configDir), 0)

	// 自动加上最后的斜杠/
	a.NotPanic(func() { initConfigDir("./testdata") })
	a.Equal(configDir, "./testdata"+string(filepath.Separator))

	a.NotPanic(func() { initConfigDir("./testdata/") })
	a.Equal(configDir, "./testdata/")
}

func TestLoadConfig(t *testing.T) {
	a := assert.New(t)

	// 指定一个不存在的配置目录，会触发panic
	a.Panic(func() { loadConfig("./emptyPath") })

	// 正常加载之后，测试各个变量是否和配置文件中的一样。
	a.NotPanic(func() { loadConfig("./testdata/web.json") })
	a.Equal(":443", cfg.Port).
		Equal("serverName", cfg.ServerName).
		True(cfg.Https).
		Equal("certFile", cfg.CertFile).
		Equal("keyFile", cfg.KeyFile)
	a.Equal(3600, cfg.Session.Lifetime).
		Equal("gosession", cfg.Session.IDName).
		Equal("memory", cfg.Session.Type).
		Equal("saveDir", cfg.Session.SaveDir)
	a.Equal("", cfg.DB["db2"].Prefix).
		Equal("dsn", cfg.DB["db2"].DSN).
		Equal("sqlite3", cfg.DB["db2"].Driver)
}
