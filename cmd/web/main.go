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

	"github.com/issue9/web/cmd/web/internal/restdoc"
)

const usageTpl = `框架 github.com/issue9/web 的辅助工具

包含了以下子命令：
{{commands}}

以及以下可用的选项：
{{flags}}

更多信息可访问 https://github.com/issue9/web 查阅。`

func main() {
	var p *localeutil.Printer
	// TODO init p

	opt := cmdopt.New(os.Stdout, flag.ContinueOnError, usageTpl, func(fs *flag.FlagSet) cmdopt.DoFunc {
		v := fs.Bool("v", false, localeutil.Phrase("show version").LocaleString(p))
		return func(w io.Writer) error {
			if *v {
				fmt.Fprintf(w, "web: %s\n", web.Version)
				fmt.Fprintf(w, "build with: %s\n", runtime.Version())
			}

			return nil
		}
	}, buildNotFound(p))

	restdoc.Init(opt, p)

	if err := opt.Exec(os.Args[1:]); err != nil {
		panic(err)
	}
}

func buildNotFound(p *localeutil.Printer) func(string) string {
	return func(s string) string {
		return localeutil.Phrase("command %s not found", s).LocaleString(p)
	}
}
