// SPDX-FileCopyrightText: 2018-2024 caixw
//
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
	"slices"
	"strings"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/source"
	"github.com/issue9/web"
	"golang.org/x/mod/modfile"

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
	depUsage       = web.StringPhrase("parse module dependencies")
	prefixUsage    = web.StringPhrase("set api path prefix")
	replaceUsage   = web.StringPhrase("parse replace direct, only valid when d is true")
)

const defaultOutput = "./restdoc.json"

func Init(opt *cmdopt.CmdOpt, p *localeutil.Printer) {
	opt.New("restdoc", title.LocaleString(p), usage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		output := fs.String("o", defaultOutput, outputUsage.LocaleString(p))
		recursive := fs.Bool("r", true, recursiveUsage.LocaleString(p))
		t := fs.String("t", "", tagUsage.LocaleString(p))
		dep := fs.Bool("d", false, depUsage.LocaleString(p))
		urlPrefix := fs.String("p", "", prefixUsage.LocaleString(p))
		replace := fs.Bool("replace", true, replaceUsage.LocaleString(p))

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

				if *dep {
					path, mod, err := source.ModFile(dir)
					if err != nil {
						return err
					}

					for _, p := range mod.Require {
						if p.Indirect {
							continue
						}

						modDir, err := getRealPath(*replace, mod, p, filepath.Dir(path))
						if err != nil {
							return err
						}
						dp.AddDir(ctx, modDir, *recursive)
					}
				}
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

var modCache = filepath.Join(build.Default.GOPATH, "pkg", "mod")

func getRealPath(replace bool, mod *modfile.File, pkg *modfile.Require, dir string) (string, error) {
	if replace && len(mod.Replace) > 0 {
		index := slices.IndexFunc(mod.Replace, func(r *modfile.Replace) bool { return r.Old.Path == pkg.Mod.Path })
		if index >= 0 {
			p := mod.Replace[index].New.Path
			if !filepath.IsAbs(p) {
				p = filepath.Join(dir, p)
			}
			return filepath.Abs(p)
		}
	}
	return filepath.Join(modCache, pkg.Mod.Path+"@"+pkg.Mod.Version), nil
}
