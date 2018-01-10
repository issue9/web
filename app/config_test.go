// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/web/internal/server"
)

var _ sanitizer = server.DefaultConfig()

func TestLoad(t *testing.T) {
	a := assert.New(t)

	conf, err := loadConfig("./testdata/web.yaml")
	a.NotError(err).NotNil(conf)
	a.Equal(conf.Root, "https://caixw.io")
	a.Equal(conf.Server.KeyFile, "keyFile")
	a.Equal(conf.Server.ReadTimeout, 3*time.Second)
}
