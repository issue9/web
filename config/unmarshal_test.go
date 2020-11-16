// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"testing"

	"github.com/issue9/assert"
)

func TestLoadYAML(t *testing.T) {
	a := assert.New(t)

	conf := &web{}
	a.NotError(LoadYAML("./testdata/web.yaml", conf))
	a.Equal(conf.Root, "http://localhost:8082")

	a.Error(LoadYAML("./testdata/web.xml", conf))
	a.ErrorIs(LoadYAML("./testdata/not-exists", conf), os.ErrNotExist)
}

func TestLoadJSON(t *testing.T) {
	a := assert.New(t)

	conf := &web{}
	a.NotError(LoadJSON("./testdata/web.json", conf))
	a.Equal(conf.Root, "http://localhost:8082")

	a.Error(LoadYAML("./testdata/web.xml", conf))
	a.ErrorIs(LoadJSON("./testdata/not-exists", conf), os.ErrNotExist)
}

func TestLoadXML(t *testing.T) {
	a := assert.New(t)

	conf := &web{}
	a.NotError(LoadXML("./testdata/web.xml", conf))
	a.Equal(conf.Root, "http://localhost:8082")

	a.Error(LoadYAML("./testdata/web.xml", conf))
	a.ErrorIs(LoadXML("./testdata/not-exists", conf), os.ErrNotExist)
}

func TestLoadFile(t *testing.T) {
	a := assert.New(t)

	xmlConf := &web{}
	yamlConf := &web{}
	jsonConf := &web{}

	a.NotError(LoadFile("./testdata/web.xml", xmlConf))
	a.NotError(LoadFile("./testdata/web.yaml", yamlConf))
	a.NotError(LoadFile("./testdata/web.json", jsonConf))
	a.Equal(xmlConf, yamlConf).Equal(yamlConf, jsonConf)

	a.Error(LoadFile("./testdata/web.not-exists", jsonConf))
}
