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
	"github.com/issue9/web/logs"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var (
	title     = localeutil.Phrase("extract locale")
	usage     = localeutil.Phrase("extract usage")
	format    = localeutil.Phrase("file format")
	out       = localeutil.Phrase("out dir")
	lang      = localeutil.Phrase("language")
	recursive = localeutil.Phrase("recursive dir")
	funcs     = localeutil.Phrase("locale func")
)

func Init(opt *cmdopt.CmdOpt, p *localeutil.Printer) {
	opt.New("locale", title.LocaleString(p), usage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		f := fs.String("f", "json", format.LocaleString(p))
		o := fs.String("o", "./locales", out.LocaleString(p))
		l := fs.String("l", "und", lang.LocaleString(p))
		r := fs.Bool("r", true, recursive.LocaleString(p))
		funcs := fs.String("func", "github.com/issue9/localeutil.Phrase,github.com/issue9/localeutil.Error", funcs.LocaleString(p))

		log, err := logs.New(&logs.Options{
			Levels:  logs.AllLevels(),
			Handler: logs.NewTermHandler(logs.NanoLayout, os.Stdout, nil),
		})
		if err != nil {
			panic(err)
		}

		return func(w io.Writer) error {
			var u message.MarshalFunc
			var ext string
			switch strings.ToLower(*f) {
			case "json", ".json":
				u = json.Marshal
				ext = ".json"
			case "yaml", "yml", ".yaml", ".yml":
				u = yaml.Marshal
				ext = ".yaml"
			default:
				return localeutil.Error("无效的参数 f")
			}

			// l
			if _, err := language.Parse(*l); err != nil {
				return err
			}

			// out
			if err := os.MkdirAll(*o, os.ModePerm); err != nil {
				return err
			}
			out := filepath.Join(*o, *l+ext)

			if fs.NArg() == 0 {
				return localeutil.Error("未指定目录")
			}

			m := &message.Messages{}
			ctx := context.Background()
			for _, dir := range fs.Args() {
				msg, err := extract.Extract(ctx, *l, dir, *r, log.ERROR(), strings.Split(*funcs, ",")...)
				if err != nil {
					return err
				}
				m.Merge(msg)
			}

			return m.SaveFile(out, u, os.ModePerm)
		}
	})
}
