// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"testing"

	"github.com/issue9/assert"
)

func TestDefaultConfig(t *testing.T) {
	a := assert.New(t)

	conf := DefaultConfig()
	a.NotError(conf.Sanitize())
}

func TestConfig_Sanitize(t *testing.T) {
	a := assert.New(t)

	conf := &Config{Port: "81"}
	a.NotError(conf.Sanitize())
	a.Equal(conf.Port, ":81")

	conf = &Config{HTTPS: true}
	a.Error(conf.Sanitize())
}
