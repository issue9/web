// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"io/fs"
	"testing"
	"time"

	"github.com/issue9/assert/v6"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"

	"github.com/issue9/web/filter"
	"github.com/issue9/web/locales"
)

var systemLanguage language.Tag

func init() {
	var err error
	systemLanguage, err = localeutil.DetectUserLanguageTag()
	if err != nil {
		panic(err)
	}
}

type empty struct{}

type userData struct {
	ID int `json:"id" yaml:"id" xml:"id,attr"`
}

func (u *userData) FilterConfig(f *filter.Filter) {
	f.Apply("id", &u.ID, filter.V2R(locales.ShouldGreatThan(0), func(id int) bool { return id > 0 }))
}

func buildFilter() (filter.Map, *filter.Filter) {
	m, c := filter.CreateMapCollector()
	return m, filter.New("", true, c)
}

func TestLoad(t *testing.T) {
	a := assert.New(t, false)

	s, data, err := Load[empty]("./testdata", "web.yaml")
	a.NotError(err).NotNil(s).Equal(data, empty{}).
		Length(s.Plugins, 3) // cache, idgen, logs

	s, data, err = Load[empty]("./testdata/not-exists", "web.yaml")
	a.ErrorIs(err, fs.ErrNotExist).Nil(s).Equal(data, empty{})
}

func TestConfig_filter(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{}
	a.Nil(conf.filter()).
		Equal(conf.languageTag, systemLanguage).
		NotNil(conf.HTTP).
		Equal(conf.location, time.Local)

	// 已经指定了 Language
	conf = &configOf[empty]{Language: "zh-hans"}
	a.Nil(conf.filter()).
		NotEqual(conf.languageTag, systemLanguage)
}

func TestConfig_buildTimezone(t *testing.T) {
	a := assert.New(t, false)
	e, f := buildFilter()

	conf := &configOf[empty]{}
	conf.buildTimezone(f)
	a.False(f.HasError()).Equal(conf.location, time.Local)

	conf = &configOf[empty]{Timezone: "Asia/Shanghai"}
	conf.buildTimezone(f)
	a.False(f.HasError()).NotNil(conf.location)

	conf = &configOf[empty]{Timezone: "undefined"}
	conf.buildTimezone(f)
	a.True(f.HasError()).
		Length(e["timezone"], 1) // 有一条错误信息
}

func TestNewPrinter(t *testing.T) {
	a := assert.New(t, false)
	p, err := NewPrinter("*.yaml", locales.Locales...)
	a.NotError(err).NotNil(p)
}
