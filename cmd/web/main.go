// SPDX-License-Identifier: MIT

//go:generate web locale -l=und -f=yaml ./
//go:generate web update-locale -src=./locales/und.yaml -dest=./locales/zh-CN.yaml

package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"runtime"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/localeutil/message/serialize"
	"github.com/issue9/web"
	wl "github.com/issue9/web/locales"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/cmd/web/build"
	"github.com/issue9/web/cmd/web/enum"
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
	p, err := newPrinter()
	if err != nil {
		panic(err)
	}

	var opt *cmdopt.CmdOpt
	opt = cmdopt.New(os.Stdout, flag.ContinueOnError, usageTpl.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		v := fs.Bool("v", false, web.StringPhrase("show version").LocaleString(p))
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

func newPrinter() (*localeutil.Printer, error) {
	tag, err := localeutil.DetectUserLanguageTag()
	if err != nil {
		fmt.Println(err)
	}

	fsys := append([]fs.FS{locales.Locales}, wl.All...)
	langs, err := serialize.LoadFSGlob(func(string) serialize.UnmarshalFunc { return yaml.Unmarshal }, "*.yaml", fsys...)
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
