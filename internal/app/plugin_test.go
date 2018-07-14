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

	a.NotError(app.loadPlugins("./plugins/*.so"))
	a.Equal(2, len(app.modules))
}

func TestApp_loadPlugin(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	m, err := app.loadPlugin("./plugins/plugin1.so")
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "plugin")

	m, err = app.loadPlugin("./plugins/not-exists.so")
	a.Error(err).Nil(m)
}
