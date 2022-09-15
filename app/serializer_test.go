// SPDX-License-Identifier: MIT

package app

import (
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
)

func TestRegisterMimetype(t *testing.T) {
	a := assert.New(t, false)

	a.NotNil(mimetypesFactory["json"].Marshal)
	RegisterMimetype(nil, nil, "json")
	a.Nil(mimetypesFactory["json"].Marshal)
}

func TestLoadConfigOf(t *testing.T) {
	a := assert.New(t, false)
	fsys := os.DirFS("./testdata")

	valid := func(conf *configOf[empty]) {
		a.Equal(conf.HTTP.Port, ":8082").
			Equal(conf.HTTP.ReadTimeout, 3*time.Second).
			Equal(conf.Timezone, "Africa/Addis_Ababa").
			Length(conf.Mimetypes, 3)
	}

	xmlConf, err := loadConfigOf[empty](fsys, "web.xml")
	a.NotError(err).NotNil(xmlConf)
	valid(xmlConf)

	yamlConf, err := loadConfigOf[empty](fsys, "web.yaml")
	a.NotError(err).NotNil(yamlConf)
	valid(yamlConf)

	jsonConf, err := loadConfigOf[empty](fsys, "web.json")
	a.NotError(err).NotNil(jsonConf)
	valid(jsonConf)

	conf, err := loadConfigOf[empty](fsys, "invalid-web.xml")
	a.Error(err).Nil(conf)
	err2, ok := err.(*ConfigError)
	a.True(ok).NotNil(err2)
	a.Equal(err2.Path, "invalid-web.xml").
		Equal(err2.Field, "http.acme.domains")

	conf, err = loadConfigOf[empty](os.DirFS("./testdata/not-exists"), "web.yaml")
	a.ErrorIs(err, fs.ErrNotExist).Nil(conf)

	customConf, err := loadConfigOf[userData](fsys, "user.xml")
	a.NotError(err).NotNil(customConf)
	a.Equal(customConf.User.ID, 1)
}
