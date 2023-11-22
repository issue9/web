// SPDX-License-Identifier: MIT

package update

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/localeutil/message"
	"github.com/issue9/localeutil/message/serialize"
	"github.com/issue9/web"
	"gopkg.in/yaml.v3"

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

			for _, d := range strings.Split(*dest, ",") {
				stat, err := os.Stat(d)
				if err != nil {
					return err
				}
				if stat.IsDir() {
					return web.NewLocaleError("the dest file %s is dir", d)
				}

				filename := filepath.Base(d)
				ext := filepath.Ext(filename)

				u, err := getUnmarshalByExt(ext)
				if err != nil {
					return err
				}

				dest, err := serialize.LoadFile(d, u)
				if err != nil {
					return err
				}

				srcMsg.MergeTo(log.WARN().LocaleString, []*message.Language{dest})

				m, _, err := locale.GetMarshalByExt(ext)
				if err != nil {
					return err
				}
				if err = serialize.SaveFile(dest, d, m, os.ModePerm); err != nil {
					return err
				}
			}

			return nil
		}
	})
}

func getSrc(src string) (*message.Language, error) {
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
