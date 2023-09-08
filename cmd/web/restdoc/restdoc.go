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
	"github.com/issue9/web/logs"

	"github.com/issue9/web/cmd/web/restdoc/logger"
	"github.com/issue9/web/cmd/web/restdoc/parser"
)

const (
	title          = localeutil.StringPhrase("gen restdoc")
	usage          = localeutil.StringPhrase("restdoc usage")
	outputUsage    = localeutil.StringPhrase("set output file")
	recursiveUsage = localeutil.StringPhrase("recursive dir")
	tagUsage       = localeutil.StringPhrase("filter by tag")
	depUsage       = localeutil.StringPhrase("parse module dependencies")
	prefixUsage    = localeutil.StringPhrase("set api path prefix")
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
			doc := parser.New(l, *urlPrefix, tags)
			for _, dir := range fs.Args() {
				doc.AddDir(ctx, dir, *r)

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
						doc.AddDir(ctx, modDir, *r)
					}
				}
			}

			return doc.Parse(ctx).SaveAs(*o)
		}
	})
}
