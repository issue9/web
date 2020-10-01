// SPDX-License-Identifier: MIT

package web

import (
	"encoding/json"
	"encoding/xml"
	"testing"
	"time"

	"github.com/issue9/assert"
	"gopkg.in/yaml.v2"
)

var (
	_ xml.Marshaler   = &Map{}
	_ xml.Unmarshaler = &Map{}

	dur time.Duration

	_ xml.Marshaler   = Duration(1)
	_ xml.Unmarshaler = (*Duration)(&dur)

	_ yaml.Marshaler   = Duration(1)
	_ yaml.Unmarshaler = (*Duration)(&dur)

	_ json.Marshaler   = Duration(1)
	_ json.Unmarshaler = (*Duration)(&dur)
)

type testMap struct {
	Pairs Map `xml:"pairs"`
}

type testDuration struct {
	Duration Duration `xml:"dur" json:"dur" yaml:"dur"`
}

func TestDuration_Duration(t *testing.T) {
	a := assert.New(t)

	dur := time.Second * 2

	a.Equal(dur, Duration(dur).Duration())
}

func TestMap(t *testing.T) {
	a := assert.New(t)

	m := &testMap{
		Pairs: Map{ // 多个字段，注意 map 顺序问题
			"key1": "val1",
		},
	}

	bs, err := xml.MarshalIndent(m, "", "  ")
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<testMap>
  <pairs>
    <key name="key1">val1</key>
  </pairs>
</testMap>`)

	rm := &testMap{}
	a.NotError(xml.Unmarshal(bs, rm))
	a.Equal(rm, m)

	// 空值
	m = &testMap{
		Pairs: Map{},
	}

	bs, err = xml.MarshalIndent(m, "", "  ")
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<testMap></testMap>`)
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

func TestDebug_sanitize(t *testing.T) {
	a := assert.New(t)

	var dbg *Debug
	a.NotError(dbg.sanitize())

	dbg = &Debug{}
	a.NotError(dbg.sanitize())

	dbg.Pprof = "abc/"
	err := dbg.sanitize()
	a.Error(err).Equal(err.Field, "pprof")

	dbg.Pprof = ""
	dbg.Vars = "abc/"
	err = dbg.sanitize()
	a.Error(err).Equal(err.Field, "vars")
}
