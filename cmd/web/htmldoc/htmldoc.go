// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package htmldoc 将对象的字段生成 html 文件
package htmldoc

import (
	"flag"
	"io"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/web"
)

const (
	title = web.StringPhrase("gen htmldoc")
	usage = web.StringPhrase("htmldoc usage")

	objectUsage = web.StringPhrase("export object name")
	dirUsae     = web.StringPhrase("set source dir")
	outputUsage = web.StringPhrase("set html doc path")
	langUsage   = web.StringPhrase("set html page language")
	titleUsage  = web.StringPhrase("set html page title")
	descUsage   = web.StringPhrase("set html page description")
	styleUsage  = web.StringPhrase("set html page stylesheet in html>head>style")
	headerUsage = web.StringPhrase("set html page header")
	footerUsage = web.StringPhrase("set html page footer")
)

const defaultStyleValue = "default"

func Init(opt *cmdopt.CmdOpt, p *localeutil.Printer) {
	opt.New("htmldoc", title.LocaleString(p), usage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		obj := fs.String("object", "", objectUsage.LocaleString(p))
		dir := fs.String("dir", "", dirUsae.LocaleString(p))
		output := fs.String("o", "", outputUsage.LocaleString(p))
		lang := fs.String("lang", "cmn-Hans", langUsage.LocaleString(p))
		title := fs.String("title", "config", titleUsage.LocaleString(p))
		desc := fs.String("desc", "", descUsage.LocaleString(p))
		style := fs.String("style", defaultStyleValue, styleUsage.LocaleString(p))
		header := fs.String("header", "", headerUsage.LocaleString(p))
		footer := fs.String("footer", "", footerUsage.LocaleString(p))

		return func(io.Writer) error {
			return export(*dir, *obj, *output, *lang, *title, *desc, *header, *footer, *style)
		}
	})
}
