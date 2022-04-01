// SPDX-License-Identifier: MIT

package app

import (
	"encoding/xml"
	"io/fs"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/logs/v4"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/serialization"
)

type empty struct{}

type userData struct {
	ID int `json:"id" yaml:"id" xml:"id,attr"`
}

func (u *userData) SanitizeConfig() *ConfigError {
	if u.ID < 0 {
		return &ConfigError{Field: "ID", Message: "必须大于 0"}
	}
	return nil
}

func TestNewOptions(t *testing.T) {
	a := assert.New(t, false)
	files := serialization.NewFiles(5)

	opt, data, err := NewOptionsOf[empty](logs.New(nil), files, os.DirFS("./testdata"), "web.yaml")
	a.Error(err).Nil(opt).Nil(data)

	a.NotError(files.Add(xml.Marshal, xml.Unmarshal, ".xml"))
	a.NotError(files.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))

	opt, data, err = NewOptionsOf[empty](logs.New(nil), files, os.DirFS("./testdata"), "web.yaml")
	a.NotError(err).NotNil(opt).Nil(data)
	a.Equal(opt.Tag, language.Und)

	opt, data, err = NewOptionsOf[empty](logs.New(nil), files, os.DirFS("./testdata/not-exists"), "web.yaml")
	a.ErrorIs(err, fs.ErrNotExist).Nil(opt).Nil(data)

	opt, data, err = NewOptionsOf[empty](logs.New(nil), files, os.DirFS("./testdata"), "invalid-web.xml")
	a.Error(err).Nil(opt).Nil(data)
	err2, ok := err.(*ConfigError)
	a.True(ok).NotNil(err2)
	a.Equal(err2.Path, "invalid-web.xml").
		Equal(err2.Field, "http.letsEncrypt.domains")

	// 自定义 T
	opt, user, err := NewOptionsOf[userData](logs.New(nil), files, os.DirFS("./testdata"), "user.xml")
	a.NotError(err).NotNil(opt).NotNil(user)
	a.Equal(user.ID, 1).Equal(opt.Port, ":8082")
}

func TestWebconfig_sanitize(t *testing.T) {
	a := assert.New(t, false)
	l := logs.New(nil)

	conf := &configOf[empty]{}
	a.NotError(conf.sanitize(l)).
		Equal(conf.languageTag, language.Und).
		NotNil(conf.HTTP).
		Nil(conf.location)

	conf = &configOf[empty]{Language: "zh-hans"}
	a.NotError(conf.sanitize(l)).
		NotEqual(conf.languageTag, language.Und)
}

func TestWebconfig_buildTimezone(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{}
	a.NotError(conf.buildTimezone()).Nil(conf.location)

	conf = &configOf[empty]{Timezone: "Asia/Shanghai"}
	a.NotError(conf.buildTimezone()).NotNil(conf.location)

	conf = &configOf[empty]{Timezone: "undefined"}
	err := conf.buildTimezone()
	a.NotNil(err).Equal(err.Field, "timezone")
}
