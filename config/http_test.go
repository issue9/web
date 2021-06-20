// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert"
)

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

	http := &HTTP{}
	http.ReadTimeout = -1
	ferr := http.sanitize()
	a.Equal(ferr.Field, "readTimeout")

	http.ReadTimeout = 0
	http.IdleTimeout = -1
	ferr = http.sanitize()
	a.Equal(ferr.Field, "idleTimeout")

	http.IdleTimeout = 0
	http.ReadHeaderTimeout = -1
	ferr = http.sanitize()
	a.Equal(ferr.Field, "readHeaderTimeout")
}

func TestHTTP_buildTLSConfig(t *testing.T) {
	a := assert.New(t)

	http := &HTTP{
		Certificates: []*Certificate{
			{
				Cert: "./testdata/cert.pem",
				Key:  "./testdata/key.pem",
			},
		},
	}
	a.NotError(http.buildTLSConfig()).NotNil(http.tlsConfig)
	a.Equal(1, len(http.tlsConfig.Certificates))

	http = &HTTP{
		LetsEncrypt: &LetsEncrypt{},
	}
	a.Error(http.buildTLSConfig()).Nil(http.tlsConfig)

	http = &HTTP{
		LetsEncrypt: &LetsEncrypt{Cache: ".", Domains: []string{"example.com"}},
	}
	a.NotError(http.buildTLSConfig()).NotNil(http.tlsConfig)

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
	a.Error(http.buildTLSConfig())
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
