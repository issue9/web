// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

//go:generate web locale -l=und -f=yaml ./
//go:generate web update-locale -src=./locales/und.yaml -dest=./locales/cmn-Hans.yaml

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/build"
	"github.com/issue9/web/cmd/web/enum"
	"github.com/issue9/web/cmd/web/htmldoc"
	"github.com/issue9/web/cmd/web/locale"
	"github.com/issue9/web/cmd/web/locale/update"
	"github.com/issue9/web/cmd/web/locales"
	"github.com/issue9/web/cmd/web/restdoc"
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

var (
	version = web.Version
	commits = ""
)

func init() {
	if commits != "" {
		version += "+" + commits
	}
}

func main() {
	p, err := locales.NewPrinter(localeutil.DetectUserLanguage())
	if err != nil {
		panic(err)
	}

	var opt *cmdopt.CmdOpt
	opt = cmdopt.New(os.Stdout, flag.ContinueOnError, usageTpl.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		v := fs.Bool("v", false, web.StringPhrase("show version").LocaleString(p))
		return func(w io.Writer) error {
			if *v {
				_, err1 := fmt.Fprintf(w, "web: %s\n", version)
				_, err2 := fmt.Fprintf(w, "build with: %s\n", runtime.Version())
				return errors.Join(err1, err2)
			}

			// 没有任何选项指定，输出帮助信息。
			_, err := io.WriteString(w, opt.Usage())
			return err
		}
	}, buildNotFound(p))

	htmldoc.Init(opt, p)
	restdoc.Init(opt, p)
	build.Init(opt, p)
	locale.Init(opt, p)
	update.Init(opt, p)
	watch.Init(opt, p)
	enum.Init(opt, p)
	cmdopt.Help(opt, "help", helpTitle.LocaleString(p), helpUsage.LocaleString(p))

	if err := opt.Exec(os.Args[1:]); err != nil {
		panic(err)
	}
}

func buildNotFound(p *localeutil.Printer) func(string) string {
	return func(s string) string {
		return web.Phrase("command %s not found", s).LocaleString(p)
	}
}
