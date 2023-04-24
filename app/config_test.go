// SPDX-License-Identifier: MIT

package app

import (
	"io/fs"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/config"
	"golang.org/x/text/language"
)

var _ config.Sanitizer = &configOf[empty]{}

type empty struct{}

type userData struct {
	ID int `json:"id" yaml:"id" xml:"id,attr"`
}

func (u *userData) SanitizeConfig() *config.FieldError {
	if u.ID < 0 {
		return config.NewFieldError("ID", "必须大于 0")
	}
	return nil
}

func TestNewServerOf(t *testing.T) {
	a := assert.New(t, false)
	const name = "app"
	const ver = "1.0"

	s, data, err := NewServerOf[empty](name, ver, "./testdata", "web.yaml")
	a.NotError(err).NotNil(s).Nil(data)

	s, data, err = NewServerOf[empty](name, ver, "./testdata/not-exists", "web.yaml")
	a.ErrorIs(err, fs.ErrNotExist).Nil(s).Nil(data)
}

func TestConfig_sanitize(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{}
	a.NotError(conf.SanitizeConfig()).
		Equal(conf.languageTag, language.Und).
		NotNil(conf.HTTP).
		Nil(conf.location)

	conf = &configOf[empty]{Language: "zh-hans"}
	a.NotError(conf.SanitizeConfig()).
		NotEqual(conf.languageTag, language.Und)
}

func TestConfig_buildTimezone(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{}
	a.NotError(conf.buildTimezone()).Nil(conf.location)

	conf = &configOf[empty]{Timezone: "Asia/Shanghai"}
	a.NotError(conf.buildTimezone()).NotNil(conf.location)

	conf = &configOf[empty]{Timezone: "undefined"}
	err := conf.buildTimezone()
	a.NotNil(err).Equal(err.Field, "timezone")
}
