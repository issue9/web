// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"encoding"
	"encoding/json/v2"
	"encoding/xml"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
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

func TestHTTP_sanitize(t *testing.T) {
	a := assert.New(t, false)
	e, f := buildFilter()

	http := &httpConfig{}
	http.ReadTimeout = -1
	http.sanitize(f)
	a.NotEmpty(e["readTimeout"])

	e, f = buildFilter()
	http.ReadTimeout = 0
	http.IdleTimeout = -1
	http.sanitize(f)
	a.NotEmpty(e["idleTimeout"])

	e, f = buildFilter()
	http.IdleTimeout = 0
	http.ReadHeaderTimeout = -1
	http.sanitize(f)
	a.NotEmpty(e["readHeaderTimeout"])
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

var (
	dur time.Duration

	_ encoding.TextMarshaler   = Duration(1)
	_ encoding.TextUnmarshaler = (*Duration)(&dur)
)

type testDuration struct {
	Duration Duration `xml:"dur" json:"dur" yaml:"dur"`
}

func TestDuration_Duration(t *testing.T) {
	a := assert.New(t, false)

	dur := time.Second * 2

	a.Equal(dur, Duration(dur).Duration())
}

func TestDuration_YAML(t *testing.T) {
	a := assert.New(t, false)

	m := &testDuration{
		Duration: Duration(time.Nanosecond * 5),
	}

	bs, err := yaml.Marshal(m)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `dur: 5ns
`)

	rm := &testDuration{}
	a.NotError(yaml.Unmarshal(bs, rm))
	a.Equal(rm, m)
}

func TestDuration_XML(t *testing.T) {
	a := assert.New(t, false)

	m := &testDuration{
		Duration: Duration(time.Nanosecond * 5),
	}

	bs, err := xml.MarshalIndent(m, "", "  ")
	a.NotError(err).NotNil(bs).
		Equal(string(bs), `<testDuration>
  <dur>5ns</dur>
</testDuration>`)

	rm := &testDuration{}
	a.NotError(xml.Unmarshal(bs, rm)).Equal(rm, m)
}

func TestDuration_XMLAttr(t *testing.T) {
	a := assert.New(t, false)

	type obj struct {
		D Duration `xml:"d,attr"`
	}
	m := &obj{
		D: Duration(time.Nanosecond * 5),
	}

	bs, err := xml.MarshalIndent(m, "", "  ")
	a.NotError(err).NotNil(bs).Equal(string(bs), `<obj d="5ns"></obj>`)

	rm := &obj{}
	a.NotError(xml.Unmarshal(bs, rm)).Equal(rm, m)
}

func TestDuration_JSON(t *testing.T) {
	a := assert.New(t, false)

	m := &testDuration{
		Duration: Duration(time.Nanosecond * 5),
	}

	bs, err := json.Marshal(m)
	a.NotError(err).NotNil(bs).Equal(string(bs), `{"dur":"5ns"}`)

	rm := &testDuration{}
	a.NotError(json.Unmarshal(bs, rm)).Equal(rm, m)
}
