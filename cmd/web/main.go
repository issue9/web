// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

//go:generate gomsgs extract -l=und -m -f=yaml -tag=comment -func=github.com/issue9/localeutil.Phrase,github.com/issue9/localeutil.Error,github.com/issue9/localeutil.StringPhrase,github.com/issue9/web.Phrase,github.com/issue9/web.StringPhrase,github.com/issue9/web.NewLocaleError,github.com/issue9/web.Context.Sprintf,github.com/issue9/web.Locale.Sprintf,github.com/issue9/web/openapi.Document.ParameterizedDoc ./
//go:generate gomsgs update -src=./locales/und.yaml -dest=./locales/zh.yaml

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/enum"
	"github.com/issue9/web/cmd/web/locales"
	"github.com/issue9/web/cmd/web/mdoc"
	"github.com/issue9/web/cmd/web/tpl"
	"github.com/issue9/web/cmd/web/watch"
)

const (
	helpTitle = web.StringPhrase("show help")
	helpUsage = web.StringPhrase("show current help info")
	usageTpl  = web.StringPhrase(`Auxiliary tool for github.com/issue9/web

commands：
{{commands}}

flags：
{{flags}}

visit https://github.com/issue9/web for more info.
`)
)

func main() {
	p, err := locales.NewPrinter(localeutil.DetectUserLanguage())
	if err != nil {
		panic(err)
	}

	var opt *cmdopt.CmdOpt
	opt = cmdopt.New(&cmdopt.Options{
		Output:        os.Stdout,
		ErrorHandling: flag.ContinueOnError,
		UsageTemplate: usageTpl.LocaleString(p),
		NotFound:      buildNotFound(p),
		Command: func(fs *flag.FlagSet) cmdopt.DoFunc {
			v := fs.Bool("v", false, web.StringPhrase("show version").LocaleString(p))

			return func(w io.Writer) error {
				if *v {
					_, err := fmt.Fprintf(w, "web: %s\nbuild with: %s\n", web.GetAppVersion(web.Version), runtime.Version())
					return err
				}

				// 没有任何选项指定，输出帮助信息。
				_, err := io.WriteString(w, opt.Usage())
				return err
			}
		},
	})

	mdoc.Init(opt, p)
	watch.Init(opt, p)
	enum.Init(opt, p)
	tpl.Init(opt, p)
	cmdopt.Help(opt, "help", helpTitle.LocaleString(p), helpUsage.LocaleString(p))

	if err := opt.Exec(os.Args[1:]); err != nil {
		panic(web.SprintError(p, true, err))
	}
}

func buildNotFound(p *localeutil.Printer) cmdopt.NotFound {
	return func(o *cmdopt.CmdOpt, s string, suggest []string) string {
		return web.Phrase("command %s not found", s).LocaleString(p)
	}
}
