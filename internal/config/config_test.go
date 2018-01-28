// Copyright 2016 by caixw, All rights reserved.
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

func TestConfig_buildRoot(t *testing.T) {
	a := assert.New(t)

	conf := &Config{}
	a.NotError(conf.buildRoot())
	a.Equal(conf.Root, "")

	conf = &Config{Root: "/"}
	a.NotError(conf.buildRoot())
	a.Equal(conf.Root, "")

	conf = &Config{Root: "/path"}
	a.NotError(conf.buildRoot())
	a.Equal(conf.Root, "/path")

	conf = &Config{Root: "/path/"}
	a.Error(conf.buildRoot())

	conf = &Config{Root: "path"}
	a.Error(conf.buildRoot())
}

func TestConfig_buildAllowedDomains(t *testing.T) {
	a := assert.New(t)

	conf := &Config{}
	a.NotError(conf.buildAllowedDomains())
	a.Empty(conf.AllowedDomains)

	// 未指定 allowedDomains
	conf.Domain = "example.com"
	a.NotError(conf.buildAllowedDomains())
	a.Empty(conf.AllowedDomains)

	// 与 domain 同一个域名
	conf.Domain = "example.com"
	conf.AllowedDomains = []string{"example.com"}
	a.NotError(conf.buildAllowedDomains())
	a.Equal(1, len(conf.AllowedDomains))

	conf.Domain = localhostURL
	conf.AllowedDomains = []string{"example.com"}
	a.NotError(conf.buildAllowedDomains())
	a.Equal(2, len(conf.AllowedDomains))

	conf.Domain = ""
	conf.AllowedDomains = []string{"example.com"}
	a.NotError(conf.buildAllowedDomains())
	a.Equal(1, len(conf.AllowedDomains))

	conf.AllowedDomains = []string{"not url"}
	a.Error(conf.buildAllowedDomains())
}

func TestConfig_buildHTTPS(t *testing.T) {
	a := assert.New(t)

	conf := &Config{HTTPS: false}
	a.NotError(conf.buildHTTPS())
	a.False(conf.HTTPS).Empty(conf.CertFile).Equal(conf.Port, 80)

	// 指定端口
	conf.Port = 8080
	a.NotError(conf.buildHTTPS())
	a.False(conf.HTTPS).Empty(conf.CertFile).Equal(conf.Port, 8080)

	// 未指定 cert 和 key
	conf.HTTPS = true
	a.Error(conf.buildHTTPS())

	conf = &Config{
		HTTPS:    true,
		CertFile: "./testdata/cert.pem",
		KeyFile:  "./testdata/key.pem",
	}
	a.NotError(conf.buildHTTPS())
	a.True(conf.HTTPS).NotEmpty(conf.CertFile).Equal(conf.Port, 443)

	// 指定端口
	conf.Port = 8080
	a.NotError(conf.buildHTTPS())
	a.True(conf.HTTPS).NotEmpty(conf.CertFile).Equal(conf.Port, 8080)
}

func TestConfig_buildURL(t *testing.T) {
	a := assert.New(t)
	conf := &Config{Port: 80}
	conf.buildURL()
	a.Equal(conf.URL, "")

	conf.Root = "/path"
	conf.URL = "" // 重置为空
	conf.buildURL()
	a.Equal(conf.URL, "/path")

	conf.Domain = localhostURL
	conf.URL = "" // 重置为空
	conf.buildURL()
	a.Equal(conf.URL, "http://localhost/path")

	conf.Port = 443
	conf.URL = "" // 重置为空
	conf.buildURL()
	a.Equal(conf.URL, "http://localhost:443/path")

	conf.Port = 80
	conf.HTTPS = true
	conf.URL = "" // 重置为空
	conf.buildURL()
	a.Equal(conf.URL, "https://localhost:80/path")

	conf.Port = 443
	conf.HTTPS = true
	conf.URL = "" // 重置为空
	conf.buildURL()
	a.Equal(conf.URL, "https://localhost/path")
}
