// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package locales 本地化内容
package locales

import (
	"embed"
	"fmt"
	"io/fs"

	gobuild "github.com/caixw/gobuild/locales"
	"github.com/issue9/localeutil"
	"github.com/issue9/localeutil/message/serialize"
	web "github.com/issue9/web/locales"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v3"
)

//go:embed *.yaml
var locales embed.FS

var Locales = append([]fs.FS{
	locales,
	gobuild.Locales,
}, web.Locales...)

func NewPrinter(lang string) (*localeutil.Printer, error) {
	tag, err := language.Parse(lang)
	if err != nil {
		fmt.Println(err)
	}

	langs, err := serialize.LoadFSGlob(func(string) serialize.UnmarshalFunc { return yaml.Unmarshal }, "*.yaml", Locales...)
	if err != nil {
		return nil, err
	}

	b := catalog.NewBuilder(catalog.Fallback(tag))
	for _, lang := range langs {
		if err := lang.Catalog(b); err != nil {
			return nil, err
		}
	}

	return message.NewPrinter(tag, message.Catalog(b)), nil
}
