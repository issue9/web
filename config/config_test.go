// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/issue9/assert"
	yaml "gopkg.in/yaml.v2"
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

func TestAddUnmarshal(t *testing.T) {
	a := assert.New(t)

	a.Error(AddUnmarshal(json.Unmarshal, "json"))

	a.NotError(AddUnmarshal(json.Unmarshal, "test1"))
	a.Error(AddUnmarshal(json.Unmarshal, ".test1"))
	a.Equal(unmarshals[".test1"], UnmarshalFunc(json.Unmarshal))

	a.NotError(AddUnmarshal(json.Unmarshal, ".test2"))
	a.Equal(unmarshals[".test2"], UnmarshalFunc(json.Unmarshal))

	a.Error(AddUnmarshal(json.Unmarshal, ""))
	a.Error(AddUnmarshal(json.Unmarshal, "."))
}

func TestSetUnmarshal(t *testing.T) {
	a := assert.New(t)

	a.NotError(SetUnmarshal(yaml.Unmarshal, "json"))
	a.Equal(unmarshals[".json"], UnmarshalFunc(yaml.Unmarshal))

	// 修改
	a.NotError(SetUnmarshal(json.Unmarshal, "test1"))
	a.NotError(SetUnmarshal(yaml.Unmarshal, ".test1"))
	a.Equal(unmarshals[".test1"], UnmarshalFunc(yaml.Unmarshal))

	// 新增
	a.NotError(SetUnmarshal(json.Unmarshal, ".test2"))
	a.Equal(unmarshals[".test2"], UnmarshalFunc(json.Unmarshal))

	a.Error(SetUnmarshal(json.Unmarshal, ""))
	a.Error(SetUnmarshal(json.Unmarshal, "."))
}

func TestLoadFile(t *testing.T) {
	a := assert.New(t)

	conf := &config{}
	a.NotError(LoadFile("./testdata/config.yaml", conf))
	a.True(conf.Debug)
	a.Equal(conf.Port, 8083)
	a.Equal(conf.CertFile, "certFile")
	a.Equal(conf.ReadTimeout, time.Second*3)

	conf = &config{}
	a.NotError(LoadFile("./testdata/config.json", conf))
	a.True(conf.Debug)
	a.Equal(conf.Port, 8083)
	a.Equal(conf.CertFile, "certFile")
	a.Equal(conf.ReadTimeout, time.Second*3)

	conf = &config{}
	a.NotError(LoadFile("./testdata/config.xml", conf))
	a.True(conf.Debug)
	a.Equal(conf.Port, 8083)
	a.Equal(conf.CertFile, "certFile")
	a.Equal(conf.ReadTimeout, time.Second*3)

	conf = &config{}
	a.Error(LoadFile("not-exists.json", conf))

	conf = &config{}
	a.Error(LoadFile("config.unknown", conf))
}
