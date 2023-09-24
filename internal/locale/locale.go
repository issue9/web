// SPDX-License-Identifier: MIT

// Package locale 用于处理本地化文件的加载
package locale

import (
	"io/fs"
	"path/filepath"

	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/localeutil/message/serialize"
	"golang.org/x/text/message/catalog"
)

// Load 通过 s 加载 fsys 中的语言文件附加在 b 之上
func Load(s config.Serializer, b *catalog.Builder, fsys fs.FS, glob string) error {
	matches, err := fs.Glob(fsys, glob)
	if err != nil {
		return err
	}

	for _, m := range matches {
		_, u := s.GetByFilename(m)
		if u == nil {
			return localeutil.Error("not found serialization function for %s", m)
		}

		l, err := serialize.LoadFS(fsys, m, u)
		if err != nil {
			return err
		}
		if err := l.Catalog(b); err != nil {
			return err
		}
	}

	return nil
}

// LoadGlob 通过 s 加载 fsys 中的语言文件附加在 b 之上
func LoadGlob(s config.Serializer, b *catalog.Builder, glob string) error {
	matches, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	for _, m := range matches {
		_, u := s.GetByFilename(m)
		if u == nil {
			return localeutil.Error("not found serialization function for %s", m)
		}

		l, err := serialize.LoadFile(m, u)
		if err != nil {
			return err
		}
		if err := l.Catalog(b); err != nil {
			return err
		}
	}

	return nil
}
