// SPDX-License-Identifier: MIT

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

	"github.com/issue9/web/cmd/web/internal/build"
	"github.com/issue9/web/cmd/web/internal/locale"
	"github.com/issue9/web/cmd/web/internal/locale/update"
	"github.com/issue9/web/cmd/web/internal/restdoc"
)

var (
	helpTitle = localeutil.Phrase("show help")
	helpUsage = localeutil.Phrase("show current help info")
	usageTpl  = localeutil.Phrase(`Auxiliary tool for github.com/issue9/web

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
	p, err := newPrinter()
	if err != nil {
		panic(err)
	}

	var opt *cmdopt.CmdOpt

	opt = cmdopt.New(os.Stdout, flag.ContinueOnError, usageTpl.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		v := fs.Bool("v", false, localeutil.Phrase("show version").LocaleString(p))
		return func(w io.Writer) error {
			if *v {
				fmt.Fprintf(w, "web: %s\n", version)
				fmt.Fprintf(w, "build with: %s\n", runtime.Version())
				return nil
			}

			// 没有任何选项指定，输出帮助信息。
			_, err := io.WriteString(w, opt.Usage())
			return err
		}
	}, buildNotFound(p))

	restdoc.Init(opt, p)
	build.Init(opt, p)
	locale.Init(opt, p)
	update.Init(opt, p)
	cmdopt.Help(opt, "help", helpTitle.LocaleString(p), helpUsage.LocaleString(p))

	if err := opt.Exec(os.Args[1:]); err != nil {
		panic(err)
	}
}

func buildNotFound(p *localeutil.Printer) func(string) string {
	return func(s string) string {
		return localeutil.Phrase("command %s not found", s).LocaleString(p)
	}
}