// SPDX-License-Identifier: MIT

// Package restdoc 生成 RESTful api 文档
package restdoc

import (
	"context"
	"flag"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/source"
	"github.com/issue9/web"
	"github.com/issue9/web/logs"

	"github.com/issue9/web/cmd/web/restdoc/logger"
	"github.com/issue9/web/cmd/web/restdoc/parser"
)

const (
	title          = web.StringPhrase("gen restdoc")
	usage          = web.StringPhrase("restdoc usage")
	outputUsage    = web.StringPhrase("set output file")
	recursiveUsage = web.StringPhrase("recursive dir")
	tagUsage       = web.StringPhrase("filter by tag")
	depUsage       = web.StringPhrase("parse module dependencies")
	prefixUsage    = web.StringPhrase("set api path prefix")
)

const defaultOutput = "./restdoc.yaml"

func Init(opt *cmdopt.CmdOpt, p *localeutil.Printer) {
	opt.New("doc", title.LocaleString(p), usage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		o := fs.String("o", defaultOutput, outputUsage.LocaleString(p))
		r := fs.Bool("r", true, recursiveUsage.LocaleString(p))
		t := fs.String("t", "", tagUsage.LocaleString(p))
		d := fs.Bool("d", false, depUsage.LocaleString(p))
		urlPrefix := fs.String("p", "", prefixUsage.LocaleString(p))

		return func(w io.Writer) error {
			ctx := context.Background()
			ls, err := logs.New(&logs.Options{
				Levels:  logs.AllLevels(),
				Handler: logs.NewTermHandler(logs.NanoLayout, os.Stdout, nil),
			})
			if err != nil {
				return err
			}

			var tags []string
			if *t != "" {
				tags = strings.Split(*t, ",")
			}

			l := logger.New(ls, p)
			dp := parser.New(l, *urlPrefix, tags)
			for _, dir := range fs.Args() {
				dp.AddDir(ctx, dir, *r)

				if *d {
					modCache := filepath.Join(build.Default.GOPATH, "pkg", "mod")

					mod, err := source.ModFile(dir)
					if err != nil {
						return err
					}

					for _, p := range mod.Require {
						if p.Indirect {
							continue
						}
						modDir := filepath.Join(modCache, p.Mod.Path+"@"+p.Mod.Version)
						dp.AddDir(ctx, modDir, *r)
					}
				}
			}

			if doc := dp.Parse(ctx); doc != nil {
				return doc.SaveAs(*o)
			}
			return nil
		}
	})
}
