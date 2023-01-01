// SPDX-License-Identifier: MIT

// Package base server 的基础环境
package base

import (
	"time"

	"github.com/issue9/logs/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
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
	tag, _, _ = l.Catalog.Matcher().Match(tag)
	return message.NewPrinter(tag, message.Catalog(l.Catalog))
}
