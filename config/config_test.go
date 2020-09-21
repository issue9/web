// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert"
)

type web struct {
	XMLName  struct{} `yaml:"-" json:"-" xml:"webconfig"`
	Debug    bool     `yaml:"debug,omitempty" json:"debug,omitempty" xml:"debug,attr,omitempty"`
	Root     string   `yaml:"root,omitempty" json:"root,omitempty" xml:"root,omitempty"`
	Timezone string   `yaml:"timezone,omitempty" json:"timezone,omitempty" xml:"timezone,omitempty"`
	Count    int      `yaml:"-" json:"-" xml:"-"`
}

func TestConfig(t *testing.T) {
	a := assert.New(t)

	w := &web{}
	var cnt int
	cfg := New("memory", w, func(config string, v interface{}) error {
		v.(*web).Count = cnt
		return nil
	}, func() { cnt++ })

	a.NotError(cfg.Refresh())
	a.Equal(w.Count, 0).
		Equal(1, cnt)

	a.NotError(cfg.Refresh())
	a.Equal(w.Count, 1).
		Equal(2, cnt)
}
