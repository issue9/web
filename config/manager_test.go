// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"testing"
	"testing/iotest"
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
	if conf.Port < 100 {
		return &Error{
			Field:   "port",
			Message: "invalid",
		}
	}

	conf.Port++
	return nil
}

func TestAddUnmarshal(t *testing.T) {
	a := assert.New(t)
	mgr, err := NewManager("./testdata")
	a.NotError(err).NotNil(mgr)

	a.NotError(mgr.AddUnmarshal(json.Unmarshal, "json"))

	a.NotError(mgr.AddUnmarshal(json.Unmarshal, "test1"))
	a.Error(mgr.AddUnmarshal(json.Unmarshal, ".test1"))
	a.Equal(mgr.unmarshals[".test1"], UnmarshalFunc(json.Unmarshal))

	a.NotError(mgr.AddUnmarshal(json.Unmarshal, ".test2"))
	a.Equal(mgr.unmarshals[".test2"], UnmarshalFunc(json.Unmarshal))

	a.Error(mgr.AddUnmarshal(json.Unmarshal, ""))
	a.Error(mgr.AddUnmarshal(json.Unmarshal, "."))
}

func TestSetUnmarshal(t *testing.T) {
	a := assert.New(t)
	mgr, err := NewManager("./testdata")
	a.NotError(err).NotNil(mgr)

	a.NotError(mgr.SetUnmarshal(yaml.Unmarshal, "json"))
	a.Equal(mgr.unmarshals[".json"], UnmarshalFunc(yaml.Unmarshal))

	// 修改
	a.NotError(mgr.SetUnmarshal(json.Unmarshal, "test1"))
	a.NotError(mgr.SetUnmarshal(yaml.Unmarshal, ".test1"))
	a.Equal(mgr.unmarshals[".test1"], UnmarshalFunc(yaml.Unmarshal))

	// 新增
	a.NotError(mgr.SetUnmarshal(json.Unmarshal, ".test2"))
	a.Equal(mgr.unmarshals[".test2"], UnmarshalFunc(json.Unmarshal))

	a.Error(mgr.SetUnmarshal(json.Unmarshal, ""))
	a.Error(mgr.SetUnmarshal(json.Unmarshal, "."))
}

func TestLoadFile(t *testing.T) {
	a := assert.New(t)
	mgr, err := NewManager("./testdata")
	a.NotError(err).NotNil(mgr)

	a.NotError(mgr.AddUnmarshal(json.Unmarshal, "json"))
	a.NotError(mgr.AddUnmarshal(yaml.Unmarshal, "yaml", ".yml"))
	a.NotError(mgr.AddUnmarshal(xml.Unmarshal, "xml"))

	conf := &config{}
	a.NotError(mgr.LoadFile("./config.yaml", conf))
	a.True(conf.Debug)
	a.Equal(conf.Port, 8083)
	a.Equal(conf.CertFile, "certFile")
	a.Equal(conf.ReadTimeout, time.Second*3)

	conf = &config{}
	a.NotError(mgr.LoadFile("./config.json", conf))
	a.True(conf.Debug)
	a.Equal(conf.Port, 8083)
	a.Equal(conf.CertFile, "certFile")
	a.Equal(conf.ReadTimeout, time.Second*3)

	conf = &config{}
	a.NotError(mgr.LoadFile("./config.xml", conf))
	a.True(conf.Debug)
	a.Equal(conf.Port, 8083)
	a.Equal(conf.CertFile, "certFile")
	a.Equal(conf.ReadTimeout, time.Second*3)

	conf = &config{}
	a.Error(mgr.LoadFile("not-exists.json", conf))

	conf = &config{}
	a.Error(mgr.LoadFile("config.unknown", conf))

	conf = &config{}
	a.ErrorType(mgr.LoadFile("./invalid.yml", conf), &Error{})
}

func TestLoad(t *testing.T) {
	a := assert.New(t)
	mgr, err := NewManager("./testdata")
	a.NotError(err).NotNil(mgr)

	a.NotError(mgr.AddUnmarshal(json.Unmarshal, "json"))
	a.NotError(mgr.AddUnmarshal(yaml.Unmarshal, "yaml", ".yml"))
	a.NotError(mgr.AddUnmarshal(xml.Unmarshal, "xml"))

	conf := &config{}
	r := bytes.NewBufferString("xx")

	errReader := iotest.DataErrReader(r)
	a.Error(mgr.Load(errReader, "json", conf))

	a.Error(mgr.Load(r, "not-exists", conf))
}
