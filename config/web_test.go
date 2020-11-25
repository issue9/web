// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"encoding/xml"
	"net/url"
	"testing"
	"time"

	"github.com/issue9/assert"
	"gopkg.in/yaml.v2"
)

var (
	dur time.Duration

	_ xml.Marshaler       = Duration(1)
	_ xml.Unmarshaler     = (*Duration)(&dur)
	_ xml.MarshalerAttr   = Duration(1)
	_ xml.UnmarshalerAttr = (*Duration)(&dur)

	_ yaml.Marshaler   = Duration(1)
	_ yaml.Unmarshaler = (*Duration)(&dur)

	_ json.Marshaler   = Duration(1)
	_ json.Unmarshaler = (*Duration)(&dur)
)

type testDuration struct {
	Duration Duration `xml:"dur" json:"dur" yaml:"dur"`
}

func TestWeb(t *testing.T) {
	a := assert.New(t)

	confYAML := &Web{}
	a.NotError(LoadFile("./testdata/web.yaml", confYAML))

	confJSON := &Web{}
	a.NotError(LoadFile("./testdata/web.json", confJSON))

	confXML := &Web{}
	a.NotError(LoadFile("./testdata/web.xml", confXML))

	a.Equal(confJSON, confXML)
	a.Equal(confJSON, confYAML)
}

func TestClassic(t *testing.T) {
	a := assert.New(t)

	srv, err := Classic("./testdata/logs.xml", "./testdata/web.yaml")
	a.NotError(err).
		NotNil(srv)
}

func TestWeb_sanitize(t *testing.T) {
	a := assert.New(t)

	conf := &Web{}
	a.NotError(conf.sanitize())
	a.Equal("Local", conf.Timezone).
		Equal(time.Local, conf.location)

	// 指定了 https，但是未指定 certificates
	conf = &Web{Root: "https://example.com"}
	err := conf.sanitize()
	a.Error(err)
	ferr, ok := err.(*FieldError)
	a.True(ok).Equal(ferr.Field, "http.certificates")
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

func TestWeb_parseResults(t *testing.T) {
	a := assert.New(t)
	conf := &Web{
		Results: map[int]Locale{
			4001:  {Key: "4001"},
			4002:  {Key: "4002"},
			50001: {Key: "50001"},
		},
	}

	a.NotError(conf.parseResults())
	a.Equal(conf.results, map[int]map[int]Locale{
		400: {
			4001: Locale{Key: "4001"},
			4002: Locale{Key: "4002"},
		},
		500: {50001: Locale{Key: "50001"}},
	})

	conf.Results[400] = Locale{Key: "400"}
	a.Error(conf.parseResults())
}

func TestDuration_Duration(t *testing.T) {
	a := assert.New(t)

	dur := time.Second * 2

	a.Equal(dur, Duration(dur).Duration())
}

func TestDuration_YAML(t *testing.T) {
	a := assert.New(t)

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
	a := assert.New(t)

	m := &testDuration{
		Duration: Duration(time.Nanosecond * 5),
	}

	bs, err := xml.MarshalIndent(m, "", "  ")
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<testDuration>
  <dur>5ns</dur>
</testDuration>`)

	rm := &testDuration{}
	a.NotError(xml.Unmarshal(bs, rm))
	a.Equal(rm, m)
}

func TestDuration_XMLAttr(t *testing.T) {
	a := assert.New(t)

	type obj struct {
		D Duration `xml:"d,attr"`
	}
	m := &obj{
		D: Duration(time.Nanosecond * 5),
	}

	bs, err := xml.MarshalIndent(m, "", "  ")
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<obj d="5ns"></obj>`)

	rm := &obj{}
	a.NotError(xml.Unmarshal(bs, rm))
	a.Equal(rm, m)
}

func TestDuration_JSON(t *testing.T) {
	a := assert.New(t)

	m := &testDuration{
		Duration: Duration(time.Nanosecond * 5),
	}

	bs, err := json.Marshal(m)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"dur":"5ns"}`)

	rm := &testDuration{}
	a.NotError(json.Unmarshal(bs, rm))
	a.Equal(rm, m)
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

func TestRouter_sanitize(t *testing.T) {
	a := assert.New(t)

	r := &Router{}
	a.NotError(r.sanitize())

	r.Pprof = "abc/"
	err := r.sanitize()
	a.Error(err).Equal(err.Field, "pprof")

	r.Pprof = ""
	r.Vars = "abc/"
	err = r.sanitize()
	a.Error(err).Equal(err.Field, "vars")
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
	a.NotError(http.buildTLSConfig(u))
	a.Equal(1, len(http.TLSConfig.Certificates))
}
