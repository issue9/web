// SPDX-License-Identifier: MIT

// Package base server 的基础环境
package base

import (
	"io/fs"
	"time"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/serialization"
)

type Base struct {
	Location *time.Location
	Catalog  *catalog.Builder
	Tag      language.Tag
	Printer  *message.Printer
	Logs     *logs.Logs
}

func New(logs *logs.Logs, loc *time.Location, tag language.Tag) *Base {
	l := &Base{
		Location: loc,
		Catalog:  catalog.NewBuilder(catalog.Fallback(tag)),
		Tag:      tag,
		Logs:     logs,
	}
	l.Printer = l.NewPrinter(tag)

	return l
}

func (l *Base) NewPrinter(tag language.Tag) *message.Printer {
	return message.NewPrinter(tag, message.Catalog(l.Catalog))
}

func (l *Base) LoadLocaleFiles(fsys fs.FS, glob string, f *serialization.FS) error {
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
