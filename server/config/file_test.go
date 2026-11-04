// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"io/fs"
	"testing"
	"time"

	"github.com/issue9/assert/v6"

	"github.com/issue9/web"
	"github.com/issue9/web/filter"
	"github.com/issue9/web/locales"
)

func TestLoadConfigOf(t *testing.T) {
	a := assert.New(t, false)
	configDir := "./testdata"

	valid := func(conf *configOf[empty]) {
		a.Equal(conf.HTTP.Port, ":8082").
			Equal(conf.HTTP.ReadTimeout, 3*time.Second).
			Equal(conf.Timezone, "Africa/Addis_Ababa").
			Length(conf.Mimetypes, 3)
	}

	xmlConf, err := loadConfigOf[empty](configDir, "web.xml")
	a.NotError(err).NotNil(xmlConf)
	valid(xmlConf)

	yamlConf, err := loadConfigOf[empty](configDir, "web.yaml")
	a.NotError(err).NotNil(yamlConf)
	valid(yamlConf)

	jsonConf, err := loadConfigOf[empty](configDir, "web.json")
	a.NotError(err).NotNil(jsonConf)
	valid(jsonConf)

	tomlConf, err := loadConfigOf[empty](configDir, "web.toml")
	a.NotError(err).NotNil(tomlConf)
	valid(tomlConf)

	jsoncConf, err := loadConfigOf[empty](configDir, "web.jsonc")
	a.NotError(err).NotNil(jsoncConf)
	valid(jsoncConf)

	conf, err := loadConfigOf[empty](configDir, "invalid-web.xml")
	a.Error(err).Nil(conf)
	merr := filter.Map{"o.http.acme.domains": []web.LocaleStringer{locales.CanNotBeEmpty}}
	a.Equal(err, web.NewError("file %s has following errors:\n%s", "invalid-web.xml", merr))

	conf, err = loadConfigOf[empty]("./testdata/not-exists", "web.yaml")
	a.ErrorIs(err, fs.ErrNotExist).Nil(conf)

	customConf, err := loadConfigOf[userData](configDir, "user.xml")
	a.NotError(err).NotNil(customConf)
	a.Equal(customConf.User.ID, 1)
}
