// SPDX-License-Identifier: MIT

package app

import (
	"encoding"
	"encoding/json"
	"encoding/xml"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"gopkg.in/yaml.v2"
)

func TestCertificate_sanitize(t *testing.T) {
	a := assert.New(t, false)

	cert := &Certificate{}
	a.Error(cert.sanitize())

	cert.Cert = "./testdata/cert.pem"
	a.Error(cert.sanitize())

	cert.Key = "./testdata/key.pem"
	a.NotError(cert.sanitize())
}

func TestHTTP_sanitize(t *testing.T) {
	a := assert.New(t, false)

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
	a := assert.New(t, false)

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
	a := assert.New(t, false)

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
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<testDuration>
  <dur>5ns</dur>
</testDuration>`)

	rm := &testDuration{}
	a.NotError(xml.Unmarshal(bs, rm))
	a.Equal(rm, m)
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
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<obj d="5ns"></obj>`)

	rm := &obj{}
	a.NotError(xml.Unmarshal(bs, rm))
	a.Equal(rm, m)
}

func TestDuration_JSON(t *testing.T) {
	a := assert.New(t, false)

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
