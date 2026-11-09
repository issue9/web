// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert/v6"
)

func TestCertificate_sanitize(t *testing.T) {
	a := assert.New(t, false)

	e, f := buildFilter()
	cert := &certificateConfig{}
	cert.sanitize(f)
	a.True(f.HasError()).NotEmpty(e["cert"])

	e, f = buildFilter()
	cert.Cert = "./testdata/cert.pem"
	cert.sanitize(f)
	a.True(f.HasError()).NotEmpty(e["key"])

	e, f = buildFilter()
	cert.Key = "./testdata/key.pem"
	cert.sanitize(f)
	a.False(f.HasError())
}

func TestHTTP_buildTLSConfig(t *testing.T) {
	a := assert.New(t, false)

	h := &httpConfig{
		Certificates: []*certificateConfig{
			{
				Cert: "./testdata/cert.pem",
				Key:  "./testdata/key.pem",
			},
		},
	}

	e, f := buildFilter()
	h.buildTLSConfig(f)
	a.False(f.HasError()).NotNil(h.tlsConfig)
	a.Equal(1, len(h.tlsConfig.Certificates)).Empty(e)

	e, f = buildFilter()
	h = &httpConfig{ACME: &acmeConfig{}}
	h.buildTLSConfig(f)
	a.True(f.HasError()).Nil(h.tlsConfig)

	e, f = buildFilter()
	h = &httpConfig{
		ACME: &acmeConfig{Cache: ".", Domains: []string{"example.com"}},
	}
	h.buildTLSConfig(f)
	a.False(f.HasError()).NotNil(h.tlsConfig)

	// 同时有 Certificates 和 LetsEncrypt
	e, f = buildFilter()
	h = &httpConfig{
		Certificates: []*certificateConfig{
			{
				Cert: "./testdata/cert.pem",
				Key:  "./testdata/key.pem",
			},
		},
		ACME: &acmeConfig{},
	}
	h.buildTLSConfig(f)
	a.True(f.HasError()).NotEmpty(e["acme"])
}

func TestACME_sanitize(t *testing.T) {
	a := assert.New(t, false)

	e, f := buildFilter()
	l := &acmeConfig{}
	l.sanitize(f)
	a.True(f.HasError()).NotEmpty(e["cache"])

	l = &acmeConfig{Cache: "./not-exists"}
	e, f = buildFilter()
	l.sanitize(f)
	a.True(f.HasError()).NotEmpty(e["cache"])

	// 未指定域名
	e, f = buildFilter()
	l = &acmeConfig{Cache: "./"}
	l.sanitize(f)
	a.True(f.HasError()).NotEmpty(e["domains"])

	e, f = buildFilter()
	l = &acmeConfig{Cache: "./", Domains: []string{"example.com"}}
	l.sanitize(f)
	a.False(f.HasError()).Empty(e)
}
