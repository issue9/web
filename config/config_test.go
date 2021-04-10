// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"testing"

	"github.com/issue9/assert"
)

var _ error = &Error{}

func TestLoadYAML(t *testing.T) {
	a := assert.New(t)
	fs := os.DirFS("./testdata")

	conf := &Webconfig{}
	a.NotError(LoadYAML(fs, "web.yaml", conf))
	a.Equal(conf.Root, "http://localhost:8082")

	conf = &Webconfig{}
	a.Error(LoadYAML(fs, "web.xml", conf))
	conf = &Webconfig{}
	a.ErrorIs(LoadYAML(fs, "not-exists", conf), os.ErrNotExist)
}

func TestLoadJSON(t *testing.T) {
	a := assert.New(t)
	fs := os.DirFS("./testdata")

	conf := &Webconfig{}
	a.NotError(LoadJSON(fs, "web.json", conf))

	conf = &Webconfig{}
	a.NotError(LoadJSON(fs, "web.json", conf))
	a.Equal(conf.Root, "http://localhost:8082")

	conf = &Webconfig{}
	a.Error(LoadJSON(fs, "web.xml", conf))
	a.ErrorIs(LoadJSON(fs, "not-exists", conf), os.ErrNotExist)
}

func TestLoadXML(t *testing.T) {
	a := assert.New(t)
	fs := os.DirFS("./testdata")

	conf := &Webconfig{}
	a.NotError(LoadXML(fs, "web.xml", conf))
	a.Equal(conf.Root, "http://localhost:8082")

	conf = &Webconfig{}
	a.Error(LoadXML(fs, "web.json", conf))
	a.ErrorIs(LoadXML(fs, "not-exists", conf), os.ErrNotExist)
}
