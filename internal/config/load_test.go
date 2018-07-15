// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestLoad(t *testing.T) {
	a := assert.New(t)

	conf, err := Load("./testdata/not-exists.yml")
	a.Error(err).Nil(conf)

	conf, err = Load("./testdata/web.yaml")
	a.NotError(err).NotNil(conf)
	a.Equal(conf.Port, 8082)
	a.Equal(conf.Domain, localhostURL)
	a.Equal(conf.ReadTimeout, time.Second*3)
	a.Equal(conf.WriteTimeout, 0)
}
