// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package plugin

import (
	"testing"
	"unicode"

	"github.com/issue9/assert"
)

func TestModuleInitFuncName(t *testing.T) {
	a := assert.New(t)

	a.True(unicode.IsUpper(rune(moduleInitFuncName[0])))
}

func TestLoad(t *testing.T) {
	a := assert.New(t)

	ms, err := Load("", nil)
	a.NotError(err).Nil(ms)

	ms, err = Load("./testdata/plugin-*.so", nil)
	a.Error(err).Nil(ms)

	ms, err = Load("./testdata/plugin_*.so", nil)
	a.NotError(err).NotNil(ms).
		Equal(2, len(ms))
}

func TestApp_loadPlugin(t *testing.T) {
	a := assert.New(t)

	m, err := loadPlugin("./testdata/plugin_1.so", nil)
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "plugin1")

	// 加载错误的插件
	m, err = loadPlugin("./testdata/plugin-3.so", nil)
	a.Error(err).Nil(m)

	// 不存在的插件
	m, err = loadPlugin("./testdata/not-exists.so", nil)
	a.Error(err).Nil(m)
}
