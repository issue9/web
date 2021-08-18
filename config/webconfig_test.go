// SPDX-License-Identifier: MIT

package config

import (
	"encoding"
	"encoding/json"
	"encoding/xml"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/serialization"
)

var (
	dur time.Duration

	_ encoding.TextMarshaler   = Duration(1)
	_ encoding.TextUnmarshaler = (*Duration)(&dur)
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

func TestNewOptions(t *testing.T) {
	a := assert.New(t)
	locale := serialization.NewLocale(catalog.NewBuilder(), serialization.NewFiles(5))

	opt, err := NewOptions(nil, locale, os.DirFS("./testdata"), "logs.xml", "web.yaml")
	a.Error(err).Nil(opt)

	a.NotError(locale.Files().Add(xml.Marshal, xml.Unmarshal, ".xml"))
	a.NotError(locale.Files().Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))

	opt, err = NewOptions(nil, locale, os.DirFS("./testdata"), "logs.xml", "web.yaml")
	a.NotError(err).NotNil(opt)

	opt, err = NewOptions(nil, locale, os.DirFS("./testdata/not-exists"), "logs.xml", "web.yaml")
	a.ErrorIs(err, fs.ErrNotExist).Nil(opt)
}
