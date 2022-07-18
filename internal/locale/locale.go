// SPDX-License-Identifier: MIT

// Package locale 本地化相关的功能
package locale

import (
	"io/fs"
	"time"

	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/filesystem"
)

type Locale struct {
	Location *time.Location
	Catalog  *catalog.Builder
	Tag      language.Tag
	Printer  *message.Printer
}

func New(loc *time.Location, tag language.Tag) *Locale {
	l := &Locale{
		Location: loc,
		Catalog:  catalog.NewBuilder(catalog.Fallback(tag)),
		Tag:      tag,
	}
	l.Printer = l.NewPrinter(tag)

	return l
}

func (l *Locale) NewPrinter(tag language.Tag) *message.Printer {
	return message.NewPrinter(tag, message.Catalog(l.Catalog))
}

func (l *Locale) LoadLocaleFiles(fsys fs.FS, glob string, f *filesystem.Serializer) error {
	matches, err := fs.Glob(fsys, glob)
	if err != nil {
		return err
	}

	for _, m := range matches {
		_, u := f.SearchByExt(m)
		if u == nil {
			return localeutil.Error("not found serialization function for %s", m)
		}

		if err := localeutil.LoadMessageFromFS(l.Catalog, fsys, m, u); err != nil {
			return err
		}
	}

	return nil
}
