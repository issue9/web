// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert"
)

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
