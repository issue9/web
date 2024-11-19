// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package locale 提取本地化内容
package locale

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/localeutil/message"
	"github.com/issue9/localeutil/message/extract"
	"github.com/issue9/localeutil/message/serialize"
	"github.com/issue9/web"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/cmd/web/termlog"
)

const (
	title = web.StringPhrase("extract locale")
	usage = web.StringPhrase(`extract usage

flags：
{{flags}}
`)
	format    = web.StringPhrase("file format")
	out       = web.StringPhrase("out dir")
	lang      = web.StringPhrase("language")
	recursive = web.StringPhrase("recursive dir")
	funcs     = web.StringPhrase("locale func")
	tag       = web.StringPhrase("locale struct tag")
	skipMod   = web.StringPhrase("skip sub module")
	info      = web.StringPhrase("show info log")
)

const presetFuncs = `github.com/issue9/localeutil.Phrase,github.com/issue9/localeutil.Error,github.com/issue9/localeutil.StringPhrase,github.com/issue9/web.Phrase,github.com/issue9/web.StringPhrase,github.com/issue9/web.NewLocaleError,github.com/issue9/web.Context.Sprintf,github.com/issue9/web.Locale.Sprintf`

const presetTag = "comment"

func Init(opt *cmdopt.CmdOpt, p *localeutil.Printer) {
	opt.New("locale", title.LocaleString(p), usage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		f := fs.String("f", "yaml", format.LocaleString(p))
		o := fs.String("o", "./locales", out.LocaleString(p))
		l := fs.String("l", "und", lang.LocaleString(p))
		r := fs.Bool("r", true, recursive.LocaleString(p))
		i := fs.Bool("i", false, info.LocaleString(p))
		t := fs.String("tag", presetTag, tag.LocaleString(p))
		funcs := fs.String("func", presetFuncs, funcs.LocaleString(p))
		skip := fs.Bool("m", true, skipMod.LocaleString(p))

		log := termlog.New(p, os.Stdout)

		return func(io.Writer) error {
			u, ext, err := GetMarshalByExt(*f)
			if err != nil {
				return err
			}

			// l
			lt, err := language.Parse(*l)
			if err != nil {
				return err
			}

			// out
			if err := os.MkdirAll(*o, os.ModePerm); err != nil {
				return err
			}
			out := filepath.Join(*o, *l+ext)

			if fs.NArg() == 0 {
				return web.NewLocaleError("no src dir")
			}

			file := &message.File{
				Languages: []language.Tag{lt},
			}
			ctx := context.Background()
			for _, dir := range fs.Args() {
				opt := &extract.Options{
					Language:      lt,
					Root:          dir,
					Recursive:     *r,
					SkipSubModule: *skip,
					WarnLog:       log.WARN().LocaleString,
					InfoLog:       func(localeutil.Stringer) {}, // 默认不输出提示信息
					Funcs:         strings.Split(*funcs, ","),
					Tag:           *t,
				}
				if *i {
					opt.InfoLog = log.INFO().LocaleString
				}

				lang, err := extract.Extract(ctx, opt)
				if err != nil {
					return err
				}
				file.Join(lang)
			}

			return serialize.SaveFile(file, out, u, os.ModePerm)
		}
	})
}

func GetMarshalByExt(ext string) (serialize.MarshalFunc, string, error) {
	switch strings.ToLower(ext) {
	case "json", ".json":
		return func(v any) ([]byte, error) {
			return json.MarshalIndent(v, "", "\t")
		}, ".json", nil
	case "yaml", "yml", ".yaml", ".yml":
		return yaml.Marshal, ".yaml", nil
	default:
		return nil, "", web.NewLocaleError("unsupported marshal for %s", ext)
	}
}
