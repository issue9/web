// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"io/fs"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/config"
	"golang.org/x/text/language"

	"github.com/issue9/web/locales"
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

func TestLoad(t *testing.T) {
	a := assert.New(t, false)

	s, data, err := Load[empty]("./testdata", "web.yaml")
	a.NotError(err).NotNil(s).Equal(data, empty{}).
		Length(s.Plugins, 3) // cache, idgen, logs

	s, data, err = Load[empty]("./testdata/not-exists", "web.yaml")
	a.ErrorIs(err, fs.ErrNotExist).Nil(s).Equal(data, empty{})
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

func TestNewPrinter(t *testing.T) {
	a := assert.New(t, false)
	p, err := NewPrinter("*.yaml", locales.Locales...)
	a.NotError(err).NotNil(p)
}
