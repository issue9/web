// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

type config struct {
	XMLName     struct{}      `json:"-" yaml:"-" xml:"config"`
	Debug       bool          `json:"debug" yaml:"debug" xml:"debug"`
	CertFile    string        `json:"certFile" yaml:"certFile" xml:"certFile"`
	Port        int           `json:"port" yaml:"port" xml:"port"`
	ReadTimeout time.Duration `json:"readTimeout" yaml:"readTimeout" xml:"readTimeout"`
}

func (conf *config) Sanitize() error {
	conf.Port++
	return nil
}

func TestLoad(t *testing.T) {
	a := assert.New(t)

	conf := &config{}
	a.NotError(Load("./testdata/config.yaml", conf))
	a.True(conf.Debug)
	a.Equal(conf.Port, 8083)
	a.Equal(conf.CertFile, "certFile")
	a.Equal(conf.ReadTimeout, time.Second*3)

	conf = &config{}
	a.NotError(Load("./testdata/config.json", conf))
	a.True(conf.Debug)
	a.Equal(conf.Port, 8083)
	a.Equal(conf.CertFile, "certFile")
	a.Equal(conf.ReadTimeout, time.Second*3)

	conf = &config{}
	a.NotError(Load("./testdata/config.xml", conf))
	a.True(conf.Debug)
	a.Equal(conf.Port, 8083)
	a.Equal(conf.CertFile, "certFile")
	a.Equal(conf.ReadTimeout, time.Second*3)

	conf = &config{}
	a.Error(Load("not-exists.json", conf))

	conf = &config{}
	a.Error(Load("config.unknown", conf))
}
