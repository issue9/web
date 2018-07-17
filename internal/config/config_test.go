// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestConfig_Load(t *testing.T) {
	a := assert.New(t)

	mgr, err := New("./testdata")
	a.NotError(err).NotNil(mgr)

	conf := &Web{}
	a.NotError(mgr.Load(mgr.File("web.yaml"), conf))
	a.Equal(conf.Port, 8082)
	a.Equal(conf.Domain, localhostURL)
	a.Equal(conf.ReadTimeout, time.Second*3)
	a.Equal(conf.WriteTimeout, 0)

	conf = &Web{}
	a.NotError(mgr.Load(mgr.File("web.json"), conf))
	a.Equal(conf.Port, 8082)
	a.Equal(conf.Domain, localhostURL)
	a.Equal(conf.ReadTimeout, time.Nanosecond*3)
	a.Equal(conf.WriteTimeout, 0)

	conf = &Web{}
	a.Error(mgr.Load("not-exists.json", conf))

	conf = &Web{}
	a.Error(mgr.Load("web.unknown", conf))
}
