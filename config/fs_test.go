// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"testing"

	"github.com/issue9/assert"
)

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
