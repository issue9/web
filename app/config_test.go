// SPDX-License-Identifier: MIT

package app

import (
	"io/fs"
	"os"
	"testing"

	"github.com/issue9/assert/v3"
	"golang.org/x/text/language"
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

func TestNewServerOf(t *testing.T) {
	a := assert.New(t, false)
	const name = "app"
	const ver = "1.0"

	s, data, err := NewServerOf[empty](name, ver, nil, os.DirFS("./testdata"), "web.yaml")
	a.NotError(err).NotNil(s).Nil(data)
	a.Equal(s.Mimetypes().Len(), 3)

	s, data, err = NewServerOf[empty](name, ver, nil, os.DirFS("./testdata/not-exists"), "web.yaml")
	a.ErrorIs(err, fs.ErrNotExist).Nil(s).Nil(data)

	s, data, err = NewServerOf[empty](name, ver, nil, os.DirFS("./testdata"), "invalid-web.xml")
	a.Error(err).Nil(s).Nil(data)
	err2, ok := err.(*ConfigError)
	a.True(ok).NotNil(err2)
	a.Equal(err2.Path, "invalid-web.xml").
		Equal(err2.Field, "http.acme.domains")

	// 自定义 T
	s, user, err := NewServerOf[userData](name, ver, nil, os.DirFS("./testdata"), "user.xml")
	a.NotError(err).NotNil(s).NotNil(user)
	a.Equal(user.ID, 1)
}

func TestWebconfig_sanitize(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{}
	a.NotError(conf.sanitize()).
		Equal(conf.languageTag, language.Und).
		NotNil(conf.HTTP).
		Nil(conf.location)

	conf = &configOf[empty]{Language: "zh-hans"}
	a.NotError(conf.sanitize()).
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
