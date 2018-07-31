// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"testing"
	"unicode"

	"github.com/issue9/assert"
)

func TestModuleInitFuncName(t *testing.T) {
	a := assert.New(t)

	a.True(unicode.IsUpper(rune(moduleInitFuncName[0])))
}

func TestApp_loadPlugins(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	a.Error(app.loadPlugins("./plugins/plugin-*.so"))

	a.NotError(app.loadPlugins("./plugins/plugin_*.so"))
	a.Equal(2, len(app.modules))
}

func TestApp_loadPlugin(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	m, err := app.loadPlugin("./plugins/plugin_1.so")
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "plugin1")

	// 加载错误的插件
	m, err = app.loadPlugin("./plugins/plugin-3.so")
	a.Error(err).Nil(m)

	// 不存在的插件
	m, err = app.loadPlugin("./plugins/not-exists.so")
	a.Error(err).Nil(m)
}
