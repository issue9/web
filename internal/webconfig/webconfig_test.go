// SPDX-License-Identifier: MIT

package webconfig

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/issue9/assert"
	"github.com/issue9/config"
)

var _ config.Sanitizer = &WebConfig{}

func TestWebConfig(t *testing.T) {
	a := assert.New(t)

	bs, err := ioutil.ReadFile("./testdata/web.yaml")
	a.NotError(err).NotNil(bs)
	confYAML := &WebConfig{}
	a.NotError(yaml.Unmarshal(bs, confYAML))

	bs, err = ioutil.ReadFile("./testdata/web.json")
	a.NotError(err).NotNil(bs)
	confJSON := &WebConfig{}
	a.NotError(json.Unmarshal(bs, confJSON))

	bs, err = ioutil.ReadFile("./testdata/web.xml")
	a.NotError(err).NotNil(bs)
	confXML := &WebConfig{}
	a.NotError(xml.Unmarshal(bs, confXML))

	a.Equal(confJSON, confXML)
	a.Equal(confJSON, confYAML)
}

func TestWebConfig_buildTimezone(t *testing.T) {
	a := assert.New(t)

	conf := &WebConfig{}
	a.NotError(conf.buildTimezone())
	a.Equal(conf.Location, time.Local).
		Equal(conf.Timezone, "Local")

	conf = &WebConfig{Timezone: "Africa/Addis_Ababa"}
	a.NotError(conf.buildTimezone())
	a.Equal(conf.Location.String(), "Africa/Addis_Ababa").
		Equal(conf.Timezone, "Africa/Addis_Ababa")

	conf = &WebConfig{Timezone: "not-exists-time-zone"}
	a.Error(conf.buildTimezone())
}

func TestWebConfig_checkStatic(t *testing.T) {
	a := assert.New(t)

	conf := &WebConfig{}
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

func TestWebConfig_buildRoot(t *testing.T) {
	a := assert.New(t)

	conf := &WebConfig{}
	a.NotError(conf.buildRoot())
	a.Equal(conf.Root, "")

	conf = &WebConfig{Root: "/"}
	a.NotError(conf.buildRoot())
	a.Equal(conf.Root, "")

	conf = &WebConfig{Root: "/path"}
	a.NotError(conf.buildRoot())
	a.Equal(conf.Root, "/path")

	conf = &WebConfig{Root: "/path/"}
	a.Error(conf.buildRoot())

	conf = &WebConfig{Root: "path"}
	a.Error(conf.buildRoot())
}

func TestIsURLPath(t *testing.T) {
	a := assert.New(t)

	a.True(isURLPath("/path"))
	a.False(isURLPath("path/"))
	a.False(isURLPath("/path/"))
	a.False(isURLPath("path"))
}

func TestWebConfig_buildAllowedDomains(t *testing.T) {
	a := assert.New(t)

	conf := &WebConfig{}
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

func TestWebConfig_buildHTTPS(t *testing.T) {
	a := assert.New(t)

	conf := &WebConfig{HTTPS: false}
	a.NotError(conf.buildHTTPS())
	a.False(conf.HTTPS).
		Nil(conf.Certificates).
		Equal(conf.Port, 80)

	// 指定端口
	conf.Port = 8080
	a.NotError(conf.buildHTTPS())
	a.False(conf.HTTPS).
		Nil(conf.Certificates).
		Equal(conf.Port, 8080)

	// 未指定 cert 和 key
	conf.HTTPS = true
	a.Error(conf.buildHTTPS())

	conf = &WebConfig{
		HTTPS: true,
		Certificates: []*Certificate{
			{
				Cert: "./testdata/cert.pem",
				Key:  "./testdata/key.pem",
			},
		},
	}
	a.NotError(conf.buildHTTPS())
	a.True(conf.HTTPS).
		Equal(1, len(conf.Certificates)).
		Equal(conf.Port, 443)

	// 指定端口
	conf.Port = 8080
	a.NotError(conf.buildHTTPS())
	a.True(conf.HTTPS).
		Equal(1, len(conf.Certificates)).
		Equal(conf.Port, 8080)
}

func TestWebConfig_buildURL(t *testing.T) {
	a := assert.New(t)
	conf := &WebConfig{Port: 80}
	a.NotError(conf.buildURL())
	a.Equal(conf.URL, "")

	conf.Root = "/path"
	conf.URL = "" // 重置为空
	a.NotError(conf.buildURL())
	a.Equal(conf.URL, "/path").
		Equal(conf.URLPath, "/path")

	conf.Domain = localhostURL
	conf.URL = "" // 重置为空
	a.NotError(conf.buildURL())
	a.Equal(conf.URL, "http://localhost/path").
		Equal(conf.URLPath, "/path")

	conf.Port = 443
	conf.URL = "" // 重置为空
	a.NotError(conf.buildURL())
	a.Equal(conf.URL, "http://localhost:443/path")

	conf.Port = 80
	conf.HTTPS = true
	conf.URL = "" // 重置为空
	a.NotError(conf.buildURL())
	a.Equal(conf.URL, "https://localhost:80/path")

	conf.Port = 443
	conf.HTTPS = true
	conf.URL = "" // 重置为空
	a.NotError(conf.buildURL())
	a.Equal(conf.URL, "https://localhost/path")

	// 强制指定 URL，不受其它参数的影响
	conf.URL = "https://example.com"
	conf.Port = 8082
	a.NotError(conf.buildURL())
	a.Equal(conf.URL, "https://example.com").
		Equal(conf.URLPath, "")
}
