// SPDX-License-Identifier: MIT

package config

import (
	"encoding/xml"
	"io/fs"
	"os"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v3"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/serialization"
)

func TestNewOptions(t *testing.T) {
	a := assert.New(t)
	files := serialization.NewFiles(5)

	opt, err := NewOptions(files, os.DirFS("./testdata"), "logs.xml", "web.yaml")
	a.Error(err).Nil(opt)

	a.NotError(files.Add(xml.Marshal, xml.Unmarshal, ".xml"))
	a.NotError(files.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))

	opt, err = NewOptions(files, os.DirFS("./testdata"), "logs.xml", "web.yaml")
	a.NotError(err).NotNil(opt)
	a.Equal(opt.Tag, language.Und)

	opt, err = NewOptions(files, os.DirFS("./testdata/not-exists"), "logs.xml", "web.yaml")
	a.ErrorIs(err, fs.ErrNotExist).Nil(opt)
}

func TestWebconfig_sanitize(t *testing.T) {
	a := assert.New(t)
	l, err := logs.New(nil)
	a.NotError(err).NotNil(l)

	conf := &Webconfig{}
	a.NotError(conf.sanitize(l)).
		Equal(conf.languageTag, language.Und).
		NotNil(conf.Router).
		NotNil(conf.HTTP).
		Nil(conf.location)

	conf = &Webconfig{Language: "zh-hans"}
	a.NotError(conf.sanitize(l)).NotEqual(conf.languageTag, language.Und)
}

func TestWebconfig_buildTimezone(t *testing.T) {
	a := assert.New(t)

	conf := &Webconfig{}
	a.NotError(conf.buildTimezone()).Nil(conf.location)

	conf = &Webconfig{Timezone: "Asia/Shanghai"}
	a.NotError(conf.buildTimezone()).NotNil(conf.location)

	conf = &Webconfig{Timezone: "undefined"}
	err := conf.buildTimezone()
	a.NotNil(err).Equal(err.Field, "timezone")
}
