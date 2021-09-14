// SPDX-License-Identifier: MIT

package config

import (
	"encoding/xml"
	"io/fs"
	"os"
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/serialization"
)

func TestNewOptions(t *testing.T) {
	a := assert.New(t)
	locale := serialization.NewLocale(catalog.NewBuilder(), serialization.NewFiles(5))

	opt, err := NewOptions(locale, os.DirFS("./testdata"), "logs.xml", "web.yaml")
	a.Error(err).Nil(opt)

	a.NotError(locale.Files().Add(xml.Marshal, xml.Unmarshal, ".xml"))
	a.NotError(locale.Files().Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))

	opt, err = NewOptions(locale, os.DirFS("./testdata"), "logs.xml", "web.yaml")
	a.NotError(err).NotNil(opt)

	opt, err = NewOptions(locale, os.DirFS("./testdata/not-exists"), "logs.xml", "web.yaml")
	a.ErrorIs(err, fs.ErrNotExist).Nil(opt)
}
