// SPDX-License-Identifier: MIT

package web

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/url"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/issue9/assert"
)

func TestWeb(t *testing.T) {
	a := assert.New(t)

	bs, err := ioutil.ReadFile("./testdata/web.yaml")
	a.NotError(err).NotNil(bs)
	confYAML := &Web{}
	a.NotError(yaml.Unmarshal(bs, confYAML))

	bs, err = ioutil.ReadFile("./testdata/web.json")
	a.NotError(err).NotNil(bs)
	confJSON := &Web{}
	a.NotError(json.Unmarshal(bs, confJSON))

	bs, err = ioutil.ReadFile("./testdata/web.xml")
	a.NotError(err).NotNil(bs)
	confXML := &Web{}
	a.NotError(xml.Unmarshal(bs, confXML))

	a.Equal(confJSON, confXML)
	a.Equal(confJSON, confYAML)
}

func TestWeb_buildTimezone(t *testing.T) {
	a := assert.New(t)

	conf := &Web{}
	a.NotError(conf.buildTimezone())
	a.Equal(conf.location, time.Local).
		Equal(conf.Timezone, "Local")

	conf = &Web{Timezone: "Africa/Addis_Ababa"}
	a.NotError(conf.buildTimezone())
	a.Equal(conf.location.String(), "Africa/Addis_Ababa").
		Equal(conf.Timezone, "Africa/Addis_Ababa")

	conf = &Web{Timezone: "not-exists-time-zone"}
	a.Error(conf.buildTimezone())
}

func TestWeb_checkStatic(t *testing.T) {
	a := assert.New(t)

	conf := &Web{}
	a.NotError(conf.checkStatic())

	conf.Static = map[string]string{
		"/admin": "./testdata",
	}
	a.NotError(conf.checkStatic())

	conf.Static = map[string]string{
		"/admin": "./not-exists",
	}
	a.Error(conf.checkStatic())

	conf.Static = map[string]string{
		"admin": "./testdata",
	}
	a.Error(conf.checkStatic())
}

func TestIsURLPath(t *testing.T) {
	a := assert.New(t)

	a.True(isURLPath("/path"))
	a.False(isURLPath("path/"))
	a.False(isURLPath("/path/"))
	a.False(isURLPath("path"))
}

func TestWeb_parseResults(t *testing.T) {
	a := assert.New(t)
	conf := &Web{
		Results: map[int]string{
			4001:  "4001",
			4002:  "4002",
			50001: "50001",
		},
	}

	a.NotError(conf.parseResults())
	a.Equal(conf.results, map[int]map[int]string{
		400: {
			4001: "4001",
			4002: "4002",
		},
		500: {50001: "50001"},
	})

	conf.Results[400] = "400"
	a.Error(conf.parseResults())
}

func TestWeb_buildAllowedDomains(t *testing.T) {
	a := assert.New(t)

	conf := &Web{
		url: &url.URL{},
	}
	a.NotError(conf.buildAllowedDomains())
	a.Empty(conf.AllowedDomains)

	// 未指定 allowedDomains
	conf.url.Host = "example.com"
	a.NotError(conf.buildAllowedDomains())
	a.Empty(conf.AllowedDomains)

	// 与 domain 同一个域名
	conf.url.Host = "example.com"
	conf.AllowedDomains = []string{"example.com"}
	a.NotError(conf.buildAllowedDomains())
	a.Equal(1, len(conf.AllowedDomains))

	conf.url.Host = "localhost"
	conf.AllowedDomains = []string{"example.com"}
	a.NotError(conf.buildAllowedDomains())
	a.Equal(2, len(conf.AllowedDomains))

	conf.url.Host = ""
	conf.AllowedDomains = []string{"example.com"}
	a.NotError(conf.buildAllowedDomains())
	a.Equal(1, len(conf.AllowedDomains))

	conf.AllowedDomains = []string{"*.example.com"}
	a.NotError(conf.buildAllowedDomains())
}

func TestLoadConfig(t *testing.T) {
	a := assert.New(t)

	conf, err := loadConfig("./testdata/web.yaml", "./testdata/logs.xml")
	a.NotError(err).NotNil(conf)

	a.True(conf.Debug).
		Equal(conf.Root, "http://localhost:8082")
}
