// SPDX-License-Identifier: MIT

package web

import (
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/web/config"
)

func TestConfig(t *testing.T) {
	a := assert.New(t)

	confYAML := &Config{}
	a.NotError(config.LoadFile("./testdata/web.yaml", confYAML))

	confJSON := &Config{}
	a.NotError(config.LoadFile("./testdata/web.json", confJSON))

	confXML := &Config{}
	a.NotError(config.LoadFile("./testdata/web.xml", confXML))

	a.Equal(confJSON, confXML)
	a.Equal(confJSON, confYAML)
}

func TestConfig_sanitize(t *testing.T) {
	a := assert.New(t)

	conf := &Config{}
	a.NotError(conf.sanitize())
	a.False(conf.isTLS).Equal(":80", conf.addr)
	a.Equal("Local", conf.Timezone).Equal(time.Local, conf.location)

	conf.ReadTimeout = -1
	err := conf.sanitize()
	a.Error(err)
	ferr, ok := err.(*config.FieldError)
	a.True(ok).Equal(ferr.Field, "readTimeout")

	conf.ReadTimeout = 0
	conf.ShutdownTimeout = -1
	err = conf.sanitize()
	a.Error(err)
	ferr, ok = err.(*config.FieldError)
	a.True(ok).Equal(ferr.Field, "shutdownTimeout")

	conf.ReadTimeout = 0
	conf.ShutdownTimeout = 0
	conf.ReadHeaderTimeout = -1
	err = conf.sanitize()
	a.Error(err)
	ferr, ok = err.(*config.FieldError)
	a.True(ok).Equal(ferr.Field, "readHeaderTimeout")

	// 指定了 https，但是未指定 certificates
	conf = &Config{Root: "https://example.com"}
	err = conf.sanitize()
	a.Error(err)
	ferr, ok = err.(*config.FieldError)
	a.True(ok).Equal(ferr.Field, "certificates")

	// 无效的 scheme
	conf = &Config{Root: "ftp://example.com"}
	err = conf.sanitize()
	a.Error(err)
	ferr, ok = err.(*config.FieldError)
	a.True(ok).Equal(ferr.Field, "root")
}

func TestConfig_buildTimezone(t *testing.T) {
	a := assert.New(t)

	conf := &Config{}
	a.NotError(conf.buildTimezone())
	a.Equal(conf.location, time.Local).
		Equal(conf.Timezone, "Local")

	conf = &Config{Timezone: "Africa/Addis_Ababa"}
	a.NotError(conf.buildTimezone())
	a.Equal(conf.location.String(), "Africa/Addis_Ababa").
		Equal(conf.Timezone, "Africa/Addis_Ababa")

	conf = &Config{Timezone: "not-exists-time-zone"}
	a.Error(conf.buildTimezone())
}

func TestConfig_checkStatic(t *testing.T) {
	a := assert.New(t)

	conf := &Config{}
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

func TestConfig_parseResults(t *testing.T) {
	a := assert.New(t)
	conf := &Config{
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

func TestLoadConfig(t *testing.T) {
	a := assert.New(t)

	conf, err := LoadConfig("./testdata")
	a.NotError(err).NotNil(conf)

	a.NotNil(conf.Debug).
		Equal(conf.Debug.Pprof, "/debug/pprof/").
		Equal(conf.Root, "http://localhost:8082")
}
