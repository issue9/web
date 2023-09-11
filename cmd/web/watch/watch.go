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
)

const (
	watchTitle    = localeutil.StringPhrase("watch and reload")
	watchUsage    = localeutil.StringPhrase("watch and reload usage")
	ignoreUsage   = localeutil.StringPhrase("not show ignore message")
	extsUsage     = localeutil.StringPhrase("set watch file extension")
	excludesUsage = localeutil.StringPhrase("exclude watch files")
	appArgsUsage  = localeutil.StringPhrase("app args")
	freqUsage     = localeutil.StringPhrase("watch frequency")
)

func Init(opt *cmdopt.CmdOpt, p *localeutil.Printer) {
	opt.New("watch", watchTitle.LocaleString(p), watchUsage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		i := fs.Bool("i", false, ignoreUsage.LocaleString(p))
		exts := fs.String("exts", ".go", extsUsage.LocaleString(p))
		excludes := fs.String("excludes", "", excludesUsage.LocaleString(p))
		appArgs := fs.String("app", "", appArgsUsage.LocaleString(p))
		freq := fs.String("freq", "1s", freqUsage.LocaleString(p))

		sources := map[string]string{
			watch.System: localeutil.StringPhrase("watch.sys").LocaleString(p),
			watch.Go:     localeutil.StringPhrase("watch.compiler").LocaleString(p),
			watch.App:    localeutil.StringPhrase("watch.app").LocaleString(p),
		}

		return func(w io.Writer) error {
			f, err := time.ParseDuration(*freq)
			if err != nil {
				return err
			}

			var m string
			if fs.NArg() == 0 {
				m = "./"
			} else {
				m = fs.Arg(0)
			}

			o := &watch.Options{
				MainFiles:        m,
				Exts:             strings.Split(*exts, ","),
				Excludes:         strings.Split(*excludes, ","),
				AppArgs:          *appArgs,
				WatcherFrequency: f,
			}

			watch.Watch(context.Background(), p, watch.NewConsoleLogger(*i, os.Stdout, nil, sources), o)
			return nil
		}
	})
}
