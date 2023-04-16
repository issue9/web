// SPDX-License-Identifier: MIT

// Package locale 用于处理本地化文件的加载
package locale

import (
	"io/fs"

	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/errs"
)

func Load(s config.Serializer, b *catalog.Builder, fsys fs.FS, glob string) error {
	matches, err := fs.Glob(fsys, glob)
	if err != nil {
		return err
	}

	for _, m := range matches {
		_, u := s.GetByFilename(m)
		if u == nil {
			return errs.NewLocaleError("not found serialization function for %s", m)
		}

		if err := localeutil.LoadMessageFromFS(b, fsys, m, u); err != nil {
			return err
		}
	}

	return nil
}
