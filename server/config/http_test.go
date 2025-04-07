// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"encoding"
	"encoding/json"
	"encoding/xml"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/issue9/assert/v4"
	"github.com/issue9/logs/v7"
)

func TestCertificate_sanitize(t *testing.T) {
	a := assert.New(t, false)

	cert := &certificateConfig{}
	a.Error(cert.sanitize())

	cert.Cert = "./testdata/cert.pem"
	a.Error(cert.sanitize())

	cert.Key = "./testdata/key.pem"
	a.NotError(cert.sanitize())
}

func TestHTTP_sanitize(t *testing.T) {
	a := assert.New(t, false)
	l := logs.New(logs.NewNopHandler())

	http := &httpConfig{}
	http.ReadTimeout = -1
	ferr := http.sanitize(l)
	a.Equal(ferr.Field, "readTimeout")

	http.ReadTimeout = 0
	http.IdleTimeout = -1
	ferr = http.sanitize(l)
	a.Equal(ferr.Field, "idleTimeout")

	http.IdleTimeout = 0
	http.ReadHeaderTimeout = -1
	ferr = http.sanitize(l)
	a.Equal(ferr.Field, "readHeaderTimeout")
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
	a.NotError(h.buildTLSConfig()).NotNil(h.tlsConfig)
	a.Equal(1, len(h.tlsConfig.Certificates))

	h = &httpConfig{
		ACME: &acmeConfig{},
	}
	a.Error(h.buildTLSConfig()).Nil(h.tlsConfig)

	h = &httpConfig{
		ACME: &acmeConfig{Cache: ".", Domains: []string{"example.com"}},
	}
	a.NotError(h.buildTLSConfig()).NotNil(h.tlsConfig)

	// 同时有 Certificates 和 LetsEncrypt
	h = &httpConfig{
		Certificates: []*certificateConfig{
			{
				Cert: "./testdata/cert.pem",
				Key:  "./testdata/key.pem",
			},
		},
		ACME: &acmeConfig{},
	}
	a.Error(h.buildTLSConfig())
}

func TestACME_sanitize(t *testing.T) {
	a := assert.New(t, false)

	l := &acmeConfig{}
	a.Error(l.sanitize())

	l = &acmeConfig{Cache: "./not-exists"}
	a.Error(l.sanitize())

	// 未指定域名
	l = &acmeConfig{Cache: "./"}
	a.Error(l.sanitize())

	l = &acmeConfig{Cache: "./", Domains: []string{"example.com"}}
	a.NotError(l.sanitize())
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
