// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/content"
	"github.com/issue9/web/internal/server"
)

var _ configer = server.DefaultConfig()

var _ configer = content.DefaultConfig()

func TestNew(t *testing.T) {
	a := assert.New(t)

	conf, err := New("./testdata")
	a.NotError(err).NotNil(conf)
	a.Equal(conf.Server.KeyFile, "keyFile")
}
