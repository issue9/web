// SPDX-License-Identifier: MIT

// Package locale 用于处理本地化文件的加载
package locale

import (
	"io/fs"

	"github.com/issue9/config"
	"github.com/issue9/localeutil/message/serialize"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

// Locale 实现了 web.Locale 接口
type Locale struct {
	*catalog.Builder
	id      language.Tag
	config  *config.Config
	printer *message.Printer
}

func New(id language.Tag, conf *config.Config, b *catalog.Builder) *Locale {
	if b == nil {
		b = catalog.NewBuilder(catalog.Fallback(id))
	}

	// 保证 b 中包含一条 id 语言的翻译项，
	// 这样可以始终让 Locale.Printer 的对象始终是有值的。
	b.SetString(id, "_____", "_____")

	return &Locale{
		Builder: b,
		id:      id,
		config:  conf,
		printer: NewPrinter(id, b),
	}
}

func (l *Locale) ID() language.Tag { return l.id }

func (l *Locale) LoadMessages(glob string, fsys ...fs.FS) error {
	return Load(l.config.Serializer(), l.Builder, glob, fsys...)
}

func (l *Locale) Printer() *message.Printer { return l.printer }

func (l *Locale) NewPrinter(id language.Tag) *message.Printer {
	if id == l.ID() {
		return l.Printer()
	}

	// TODO 以使用频次或是 TTL 的方式缓存常用的 Printer，
	// 可以在一定程序上提升 NewPrinter 的性能。
	return NewPrinter(id, l)
}

// NewPrinter 从 cat 是查找最符合 tag 的语言 ID 并返回对应的 [message.Printer] 对象
func NewPrinter(tag language.Tag, cat catalog.Catalog) *message.Printer {
	tag, _, _ = cat.Matcher().Match(tag) // 从 cat 中查找最合适的 tag
	return message.NewPrinter(tag, message.Catalog(cat))
}

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
