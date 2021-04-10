// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"testing"

	"github.com/issue9/assert"
)

var _ error = &Error{}

type object struct {
	XMLName  struct{} `yaml:"-" json:"-" xml:"web"`
	Root     string   `yaml:"root,omitempty" json:"root,omitempty" xml:"root,omitempty"`
	Timezone string   `yaml:"timezone,omitempty" json:"timezone,omitempty" xml:"timezone,omitempty"`
	Count    int      `yaml:"-" json:"-" xml:"-"`
}

func TestRefresher(t *testing.T) {
	a := assert.New(t)

	w := &object{}
	r, err := Load("memory", w, func(config string, v interface{}) error {
		v.(*object).Count++
		return nil
	})
	a.NotError(err).NotNil(r)

	a.NotError(r.Refresh())
	a.Equal(w.Count, 1)

	a.NotError(r.Refresh())
	a.Equal(w.Count, 1) // Refresh 会重新创建一个新对象并赋值给 w，不会增加 w.Count 的值。
}

func TestLoadYAML(t *testing.T) {
	a := assert.New(t)

	conf := &object{}
	f := LoadYAML(os.DirFS("./testdata"))
	a.NotError(f("web.yaml", conf))
	a.Equal(conf.Root, "http://localhost:8082")

	a.Error(f("web.xml", conf))
	a.ErrorIs(f("not-exists", conf), os.ErrNotExist)
}

func TestLoadJSON(t *testing.T) {
	a := assert.New(t)

	conf := &object{}
	f := LoadJSON(os.DirFS("./testdata"))
	a.NotError(f("web.json", conf))
	a.Equal(conf.Root, "http://localhost:8082")

	a.Error(f("web.xml", conf))
	a.ErrorIs(f("not-exists", conf), os.ErrNotExist)
}

func TestLoadXML(t *testing.T) {
	a := assert.New(t)

	conf := &object{}
	f := LoadXML(os.DirFS("./testdata"))
	a.NotError(f("web.xml", conf))
	a.Equal(conf.Root, "http://localhost:8082")

	a.Error(f("web.json", conf))
	a.ErrorIs(f("not-exists", conf), os.ErrNotExist)
}
