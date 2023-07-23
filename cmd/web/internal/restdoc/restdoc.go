// SPDX-License-Identifier: MIT

// Package restdoc 生成 RESTful api 文档
//
// TODO 文档
// map 无法指定字段名，转换成空对象， interface{} 则无法转换。
// 不支持 gopath 模式
package restdoc

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/web/logs"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger"
	"github.com/issue9/web/cmd/web/internal/restdoc/parser"
)

const (
	title          = localeutil.StringPhrase("gen restdoc")
	usage          = localeutil.StringPhrase("restdoc usage")
	outputUsage    = localeutil.StringPhrase("set output file")
	recursiveUsage = localeutil.StringPhrase("recursive dir")
)

const defaultOutput = "./restdoc.json"

func Init(opt *cmdopt.CmdOpt, p *localeutil.Printer) {
	opt.New("doc", title.LocaleString(p), usage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		o := fs.String("o", defaultOutput, outputUsage.LocaleString(p))
		r := fs.Bool("r", true, recursiveUsage.LocaleString(p))

		return func(w io.Writer) error {
			ctx := context.Background()
			ls, err := logs.New(&logs.Options{
				Levels:  logs.AllLevels(),
				Handler: logs.NewTermHandler(logs.NanoLayout, os.Stdout, nil),
			})
			if err != nil {
				return err
			}

			l := logger.New(ls, p)
			doc := parser.New(l)
			for _, dir := range fs.Args() {
				doc.AddDir(ctx, dir, *r)
			}

			t := doc.OpenAPI(ctx)

			data, err := json.Marshal(t)
			if err != nil {
				return err
			}
			return os.WriteFile(*o, data, os.ModePerm)
		}
	})
}
