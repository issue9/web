// SPDX-License-Identifier: MIT

package config

import (
	"encoding/xml"
	"net/url"
	"testing"

	"github.com/issue9/assert"
)

var (
	_ xml.Marshaler   = &Pairs{}
	_ xml.Unmarshaler = &Pairs{}
)

type testPairs struct {
	Pairs Pairs `xml:"pairs"`
}

func TestCertificate_sanitize(t *testing.T) {
	a := assert.New(t)

	cert := &Certificate{}
	a.Error(cert.sanitize())

	cert.Cert = "./testdata/cert.pem"
	a.Error(cert.sanitize())

	cert.Key = "./testdata/key.pem"
	a.NotError(cert.sanitize())
}

func TestHTTP_sanitize(t *testing.T) {
	a := assert.New(t)
	u, err := url.Parse("/")
	a.NotError(err).NotNil(u)

	http := &HTTP{}
	http.ReadTimeout = -1
	ferr := http.sanitize(u)
	a.Equal(ferr.Field, "readTimeout")

	http.ReadTimeout = 0
	http.IdleTimeout = -1
	ferr = http.sanitize(u)
	a.Equal(ferr.Field, "idleTimeout")

	http.IdleTimeout = 0
	http.ReadHeaderTimeout = -1
	ferr = http.sanitize(u)
	a.Equal(ferr.Field, "readHeaderTimeout")
}

func TestHTTP_buildTLSConfig(t *testing.T) {
	a := assert.New(t)
	u, err := url.Parse("/")
	a.NotError(err).NotNil(u)

	http := &HTTP{
		Certificates: []*Certificate{
			{
				Cert: "./testdata/cert.pem",
				Key:  "./testdata/key.pem",
			},
		},
	}
	a.NotError(http.buildTLSConfig(u)).NotNil(http.tlsConfig)
	a.Equal(1, len(http.tlsConfig.Certificates))

	http = &HTTP{
		LetsEncrypt: &LetsEncrypt{},
	}
	a.Error(http.buildTLSConfig(u)).Nil(http.tlsConfig)

	http = &HTTP{
		LetsEncrypt: &LetsEncrypt{Cache: ".", Domains: []string{"example.com"}},
	}
	a.NotError(http.buildTLSConfig(u)).NotNil(http.tlsConfig)

	// 同时有 Certificates 和 LetsEncrypt
	http = &HTTP{
		Certificates: []*Certificate{
			{
				Cert: "./testdata/cert.pem",
				Key:  "./testdata/key.pem",
			},
		},
		LetsEncrypt: &LetsEncrypt{},
	}
	a.Error(http.buildTLSConfig(u))
}

func TestLetEncrypt_sanitize(t *testing.T) {
	a := assert.New(t)

	l := &LetsEncrypt{}
	a.Error(l.sanitize())

	l = &LetsEncrypt{Cache: "./not-exists"}
	a.Error(l.sanitize())

	// 未指定域名
	l = &LetsEncrypt{Cache: "./"}
	a.Error(l.sanitize())

	l = &LetsEncrypt{Cache: "./", Domains: []string{"example.com"}}
	a.NotError(l.sanitize())
}

func TestPairs(t *testing.T) {
	a := assert.New(t)

	m := &testPairs{
		Pairs: Pairs{ // 多个字段，注意 map 顺序问题
			"key1": "val1",
		},
	}

	bs, err := xml.MarshalIndent(m, "", "  ")
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<testPairs>
  <pairs>
    <key name="key1">val1</key>
  </pairs>
</testPairs>`)

	rm := &testPairs{}
	a.NotError(xml.Unmarshal(bs, rm))
	a.Equal(rm, m)

	// 空值
	m = &testPairs{
		Pairs: Pairs{},
	}

	bs, err = xml.MarshalIndent(m, "", "  ")
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<testPairs></testPairs>`)
}
