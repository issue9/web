// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"testing"

	"github.com/issue9/assert"
)

func TestLoad(t *testing.T) {
	a := assert.New(t)

	conf, err := Load("./testdata/web.json")
	a.NotError(err).NotNil(conf)
}
