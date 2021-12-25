// SPDX-License-Identifier: MIT

package app

import (
	"encoding/xml"
	"io/fs"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/serialization"
)

func TestNewOptions(t *testing.T) {
	a := assert.New(t, false)
	files := serialization.NewFiles(5)

	opt, err := NewOptions(files, os.DirFS("./testdata"), "web.yaml")
	a.Error(err).Nil(opt)

	a.NotError(files.Add(xml.Marshal, xml.Unmarshal, ".xml"))
	a.NotError(files.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))

	opt, err = NewOptions(files, os.DirFS("./testdata"), "web.yaml")
	a.NotError(err).NotNil(opt)
	a.Equal(opt.Tag, language.Und)

	opt, err = NewOptions(files, os.DirFS("./testdata/not-exists"), "web.yaml")
	a.ErrorIs(err, fs.ErrNotExist).Nil(opt)

	opt, err = NewOptions(files, os.DirFS("./testdata"), "invalid-web.xml")
	a.Error(err).Nil(opt)
	err2, ok := err.(*Error)
	a.True(ok).NotNil(err2)
	a.Equal(err2.Config, "invalid-web.xml").
		Equal(err2.Field, "router.cors.allowCredentials")
}

func TestWebconfig_sanitize(t *testing.T) {
	a := assert.New(t, false)

	conf := &Webconfig{}
	a.NotError(conf.sanitize()).
		Equal(conf.languageTag, language.Und).
		NotNil(conf.Router).
		NotNil(conf.HTTP).
		Nil(conf.location)

	conf = &Webconfig{Language: "zh-hans"}
	a.NotError(conf.sanitize()).
		NotEqual(conf.languageTag, language.Und).
		NotNil(conf.logs)
}

func TestWebconfig_buildTimezone(t *testing.T) {
	a := assert.New(t, false)

	conf := &Webconfig{}
	a.NotError(conf.buildTimezone()).Nil(conf.location)

	conf = &Webconfig{Timezone: "Asia/Shanghai"}
	a.NotError(conf.buildTimezone()).NotNil(conf.location)

	conf = &Webconfig{Timezone: "undefined"}
	err := conf.buildTimezone()
	a.NotNil(err).Equal(err.Field, "timezone")
}
