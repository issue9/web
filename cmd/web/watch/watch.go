// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package watch 热编译项目
package watch

import (
	"context"
	"flag"
	"io"
	"os"
	"strings"
	"time"

	"github.com/caixw/gobuild/watch"
	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/web"
)

const (
	watchTitle    = web.StringPhrase("watch and reload")
	watchUsage    = web.StringPhrase("watch and reload usage")
	ignoreUsage   = web.StringPhrase("not show ignore message")
	extsUsage     = web.StringPhrase("set watch file extension")
	excludesUsage = web.StringPhrase("exclude watch files")
	appArgsUsage  = web.StringPhrase("app args")
	freqUsage     = web.StringPhrase("watch frequency")
	devUsage      = web.StringPhrase("dev mode")
)

func Init(opt *cmdopt.CmdOpt, p *localeutil.Printer) {
	opt.New("watch", watchTitle.LocaleString(p), watchUsage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		i := fs.Bool("i", false, ignoreUsage.LocaleString(p))
		exts := fs.String("exts", ".go", extsUsage.LocaleString(p))
		excludes := fs.String("excludes", "", excludesUsage.LocaleString(p))
		appArgs := fs.String("app", "", appArgsUsage.LocaleString(p))
		freq := fs.String("freq", "1s", freqUsage.LocaleString(p))
		dev := fs.Bool("dev", true, devUsage.LocaleString(p))

		sources := map[string]string{
			watch.System: web.StringPhrase("watch.sys").LocaleString(p),
			watch.Go:     web.StringPhrase("watch.compiler").LocaleString(p),
			watch.App:    web.StringPhrase("watch.app").LocaleString(p),
		}

		return func(io.Writer) error {
			f, err := time.ParseDuration(*freq)
			if err != nil {
				return err
			}

			var mf string
			if fs.NArg() == 0 {
				mf = "./"
			} else {
				mf = fs.Arg(0)
			}

			var args []string
			if *dev {
				args = append(args, "-tags=development")
			}

			o := &watch.Options{
				MainFiles:        mf,
				Args:             args,
				Exts:             strings.Split(*exts, ","),
				Excludes:         strings.Split(*excludes, ","),
				AppArgs:          *appArgs,
				WatcherFrequency: f,
			}

			return watch.Watch(context.Background(), p, watch.NewConsoleLogger(*i, os.Stdout, nil, sources), o)
		}
	})
}
