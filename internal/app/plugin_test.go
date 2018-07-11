// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"testing"

	"github.com/issue9/assert"
)

func TestApp_loadPlugin(t *testing.T) {
	a := assert.New(t)
	app, err := New("./testdata")
	a.NotError(err).NotNil(app)

	m, err := app.loadPlugin("./plugin/plugin.so")
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "plugin")

	m, err = app.loadPlugin("./plugin/not-exists.so")
	a.Error(err).Nil(m)
}
