// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package modules

import (
	"testing"
	"unicode"

	"github.com/issue9/assert"
)

func TestModuleInitFuncName(t *testing.T) {
	a := assert.New(t)

	a.True(unicode.IsUpper(rune(moduleInitFuncName[0])))
}

func TestLoadPlugins(t *testing.T) {
	a := assert.New(t)

	ms, err := loadPlugins("./testdata/plugin-*.so")
	a.Error(err).Nil(ms)

	ms, err = loadPlugins("./testdata/plugin_*.so")
	if !isPluginOS() {
		a.Error(err).Nil(ms)
	} else {
		a.NotError(err).NotNil(ms).
			Equal(2, len(ms))
	}
}

func TestApp_loadPlugin(t *testing.T) {
	a := assert.New(t)

	if !isPluginOS() {
		return
	}

	m, err := loadPlugin("./testdata/plugin_1.so")
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "plugin1")
	a.NotEmpty(m.Routes["/plugin1"]["GET"])
	a.Empty(m.Routes["/plugin1"]["POST"])

	// 加载错误的插件
	m, err = loadPlugin("./testdata/plugin-3.so")
	a.Error(err).Nil(m)

	// 不存在的插件
	m, err = loadPlugin("./testdata/not-exists.so")
	a.Error(err).Nil(m)
}