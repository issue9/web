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

type empty struct{}

type userData struct {
	ID int `json:"id" yaml:"id" xml:"id,attr"`
}

func (u *userData) Sanitize() *Error {
	if u.ID < 0 {
		return &Error{Field: "ID", Message: "必须大于 0"}
	}
	return nil
}

func TestNewOptions(t *testing.T) {
	a := assert.New(t, false)
	files := serialization.NewFiles(5)

	opt, data, err := NewOptions[empty](files, os.DirFS("./testdata"), "web.yaml")
	a.Error(err).Nil(opt).Nil(data)

	a.NotError(files.Add(xml.Marshal, xml.Unmarshal, ".xml"))
	a.NotError(files.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))

	opt, data, err = NewOptions[empty](files, os.DirFS("./testdata"), "web.yaml")
	a.NotError(err).NotNil(opt).Nil(data)
	a.Equal(opt.Tag, language.Und)

	opt, data, err = NewOptions[empty](files, os.DirFS("./testdata/not-exists"), "web.yaml")
	a.ErrorIs(err, fs.ErrNotExist).Nil(opt).Nil(data)

	opt, data, err = NewOptions[empty](files, os.DirFS("./testdata"), "invalid-web.xml")
	a.Error(err).Nil(opt).Nil(data)
	err2, ok := err.(*Error)
	a.True(ok).NotNil(err2)
	a.Equal(err2.Config, "invalid-web.xml").
		Equal(err2.Field, "router.cors.allowCredentials")

	// 自定义 T
	opt, user, err := NewOptions[userData](files, os.DirFS("./testdata"), "user.xml")
	a.NotError(err).NotNil(opt).NotNil(user)
	a.Equal(user.ID, 1).Equal(opt.Port, ":8082")
}

func TestWebconfig_sanitize(t *testing.T) {
	a := assert.New(t, false)

	conf := &webconfig[empty]{}
	a.NotError(conf.sanitize()).
		Equal(conf.languageTag, language.Und).
		NotNil(conf.Router).
		NotNil(conf.HTTP).
		Nil(conf.location)

	conf = &webconfig[empty]{Language: "zh-hans"}
	a.NotError(conf.sanitize()).
		NotEqual(conf.languageTag, language.Und).
		NotNil(conf.logs)
}

func TestWebconfig_buildTimezone(t *testing.T) {
	a := assert.New(t, false)

	conf := &webconfig[empty]{}
	a.NotError(conf.buildTimezone()).Nil(conf.location)

	conf = &webconfig[empty]{Timezone: "Asia/Shanghai"}
	a.NotError(conf.buildTimezone()).NotNil(conf.location)

	conf = &webconfig[empty]{Timezone: "undefined"}
	err := conf.buildTimezone()
	a.NotNil(err).Equal(err.Field, "timezone")
}
