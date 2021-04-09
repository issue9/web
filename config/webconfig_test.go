// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"encoding/xml"
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
