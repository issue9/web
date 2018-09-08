// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package modules

import (
	"testing"

	"github.com/issue9/assert"
)

func TestLoadPlugins(t *testing.T) {
	a := assert.New(t)

	ms, err := loadPlugins("./testdata/plugin-*.so", nil)
	a.Error(err).Nil(ms)

	ms, err = loadPlugins("./testdata/plugin_*.so", nil)
	a.Error(err).Nil(ms)
}
