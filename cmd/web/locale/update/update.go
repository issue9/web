// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

// Package update 更新本地化文档
package update

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/localeutil/message"
	"github.com/issue9/localeutil/message/serialize"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/locale"
	"github.com/issue9/web/cmd/web/termlog"
)

const (
	title = web.StringPhrase("update locale file")
	usage = web.StringPhrase(`update locale file usage

flags:
{{flags}}
`)
	srcUsage  = web.StringPhrase("src locale file")
	destUsage = web.StringPhrase("dest locale files")
)

func Init(opt *cmdopt.CmdOpt, p *localeutil.Printer) {
	opt.New("update-locale", title.LocaleString(p), usage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		src := fs.String("src", "", srcUsage.LocaleString(p))
		dest := fs.String("dest", "", destUsage.LocaleString(p))

		return func(io.Writer) error {
			log := termlog.New(p, os.Stdout)

			srcMsg, err := getSrc(*src)
			if err != nil {
				return err
			}

			for _, path := range strings.Split(*dest, ",") {
				var dest *message.File

				stat, err := os.Stat(path)
				if errors.Is(err, os.ErrNotExist) {
					dest = &message.File{}
				} else if err != nil {
					return err
				} else if stat.IsDir() {
					return web.NewLocaleError("the dest file %s is dir", path)
				}

				filename := filepath.Base(path)
				ext := filepath.Ext(filename)
				u, err := getUnmarshalByExt(ext)
				if err != nil {
					return err
				}
				if dest == nil { // dest != nil，说明因为不存在文件，已经被初始经默认值。
					if dest, err = serialize.LoadFile(path, u); err != nil {
						return err
					}
				}

				srcMsg.MergeTo(log.WARN().LocaleString, dest, path)

				m, _, err := locale.GetMarshalByExt(ext)
				if err != nil {
					return err
				}
				if err = serialize.SaveFile(dest, path, m, os.ModePerm); err != nil {
					return err
				}
			}

			return nil
		}
	})
}

func getSrc(src string) (*message.File, error) {
	u, err := getUnmarshalByExt(filepath.Ext(src))
	if err != nil {
		return nil, err
	}

	return serialize.LoadFile(src, u)
}

func getUnmarshalByExt(ext string) (serialize.UnmarshalFunc, error) {
	switch strings.ToLower(ext) {
	case "json", ".json":
		return json.Unmarshal, nil
	case "yaml", "yml", ".yaml", ".yml":
		return yaml.Unmarshal, nil
	default:
		return nil, web.NewLocaleError("unsupported unmarshal for %s", ext)
	}
}
