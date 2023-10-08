// SPDX-License-Identifier: MIT

// Package locale 用于处理本地化文件的加载
package locale

import (
	"io/fs"

	"github.com/issue9/config"
	"github.com/issue9/localeutil/message/serialize"
	"golang.org/x/text/message/catalog"
)

func Load(s config.Serializer, b *catalog.Builder, glob string, fsys ...fs.FS) error {
	search := func(p string) serialize.UnmarshalFunc {
		_, u := s.GetByFilename(p)
		return u
	}
	langs, err := serialize.LoadFSGlob(search, glob, fsys...)
	if err != nil {
		return err
	}

	for _, lang := range langs {
		if err := lang.Catalog(b); err != nil {
			return err
		}
	}
	return nil
}
