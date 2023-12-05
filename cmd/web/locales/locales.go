// SPDX-License-Identifier: MIT

// Package locales 本地化内容
package locales

import (
	"embed"
	"io/fs"

	gobuild "github.com/caixw/gobuild/locales"
	web "github.com/issue9/web/locales"
)

//go:embed *.yaml
var locales embed.FS

var Locales = []fs.FS{
	locales,
	gobuild.Locales,
}

func init() {
	Locales = append(Locales, web.Locales...)
}
