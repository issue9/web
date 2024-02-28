// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"io/fs"
	"path/filepath"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/config"

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

	conf, err := loadConfigOf[empty](configDir, "invalid-web.xml")
	a.Error(err).Nil(conf)
	err2, ok := err.(*config.FieldError)
	a.True(ok).NotNil(err2)
	a.Equal(err2.Path, filepath.Join("testdata", "invalid-web.xml")).
		Equal(err2.Field, "http.acme.domains")

	conf, err = loadConfigOf[empty]("./testdata/not-exists", "web.yaml")
	a.ErrorIs(err, fs.ErrNotExist).Nil(conf)

	customConf, err := loadConfigOf[userData](configDir, "user.xml")
	a.NotError(err).NotNil(customConf)
	a.Equal(customConf.User.ID, 1)
}

func TestNewPrinter(t *testing.T) {
	a := assert.New(t, false)
	p, err := NewPrinter("*.yaml", locales.Locales...)
	a.NotError(err).NotNil(p)
}
