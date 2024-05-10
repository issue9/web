// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package restdoc 生成 RESTful api 文档
package restdoc

import (
	"context"
	"flag"
	"io"
	"os"
	"strings"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/restdoc/logger"
	"github.com/issue9/web/cmd/web/restdoc/parser"
	"github.com/issue9/web/cmd/web/termlog"
)

const (
	title          = web.StringPhrase("gen restdoc")
	usage          = web.StringPhrase("restdoc usage")
	outputUsage    = web.StringPhrase("set output file")
	recursiveUsage = web.StringPhrase("recursive dir")
	tagUsage       = web.StringPhrase("filter by tag")
	prefixUsage    = web.StringPhrase("set api path prefix")
)

const defaultOutput = "./restdoc.json"

func Init(opt *cmdopt.CmdOpt, p *localeutil.Printer) {
	opt.New("restdoc", title.LocaleString(p), usage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		output := fs.String("o", defaultOutput, outputUsage.LocaleString(p))
		recursive := fs.Bool("r", true, recursiveUsage.LocaleString(p))
		t := fs.String("t", "", tagUsage.LocaleString(p))
		urlPrefix := fs.String("p", "", prefixUsage.LocaleString(p))

		return func(io.Writer) error {
			ctx := context.Background()
			l := logger.New(termlog.New(p, os.Stdout))

			go func() {
				if msg := recover(); msg != nil {
					l.Fatal(msg)
				}
			}()

			var tags []string
			if *t != "" {
				tags = strings.Split(*t, ",")
			}

			dp := parser.New(l, *urlPrefix, tags)
			for _, dir := range fs.Args() {
				dp.AddDir(ctx, dir, *recursive)
			}

			if doc := dp.Parse(ctx); doc != nil {
				if err := doc.SaveAs(*output); err != nil {
					return err
				}
				l.Info(web.NewLocaleError("save restdoc to %s", *output))
			}
			return nil
		}
	})
}
