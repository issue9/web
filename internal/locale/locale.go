// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

// Package locale 用于处理本地化文件的加载
package locale

import (
	"io/fs"

	"github.com/issue9/config"
	"github.com/issue9/localeutil/message/serialize"
	"github.com/jellydator/ttlcache/v3"
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
	ttl     *ttlcache.Cache[language.Tag, *message.Printer]

	acceptLanguage string // 根据内容生成 Accept-Language 报头
}

func New(id language.Tag, conf *config.Config) *Locale {
	b := catalog.NewBuilder(catalog.Fallback(id))

	// 保证 b 中包含一条 id 语言的翻译项，
	// 这样可以始终让 Locale.Printer 的对象始终是有值的。
	if err := b.SetString(id, "_____", "_____"); err != nil {
		panic(err)
	}

	p, _ := NewPrinter(id, b)
	return &Locale{
		Builder:        b,
		id:             id,
		config:         conf,
		printer:        p,
		ttl:            ttlcache.New(ttlcache.WithCapacity[language.Tag, *message.Printer](10)),
		acceptLanguage: buildAcceptLanguage(b),
	}
}

func buildAcceptLanguage(ls catalog.Catalog) string {
	var s string
	for index, v := range ls.Languages() {
		if index == 0 {
			s += v.String()
		} else {
			s += ", " + v.String()
		}
	}
	return s
}

func (l *Locale) ID() language.Tag { return l.id }

// AcceptLanguage 返回当前 Locale 形成的 Accept-Language 报头
func (l *Locale) AcceptLanguage() string { return l.acceptLanguage }

func (l *Locale) LoadMessages(glob string, fsys ...fs.FS) error {
	if err := Load(l.Config().Serializer(), l.Builder, glob, fsys...); err != nil {
		return err
	}

	l.acceptLanguage = buildAcceptLanguage(l.Builder)

	return nil
}

func (l *Locale) Printer() *message.Printer { return l.printer }

func (l *Locale) Config() *config.Config { return l.config }

func (l *Locale) Sprintf(key string, v ...any) string { return l.Printer().Sprintf(key, v...) }

func (l *Locale) NewPrinter(id language.Tag) *message.Printer {
	if id == l.ID() {
		return l.Printer()
	}

	if item := l.ttl.Get(id); item != nil {
		return item.Value()
	}
	p, exact := NewPrinter(id, l)
	if exact {
		l.ttl.Set(id, p, ttlcache.DefaultTTL)
	}
	return p
}

// NewPrinter 从 cat 是查找最符合 tag 的语言 ID 并返回对应的 [message.Printer] 对象
func NewPrinter(tag language.Tag, cat catalog.Catalog) (*message.Printer, bool) {
	tag, _, confidence := cat.Matcher().Match(tag) // 从 cat 中查找最合适的 tag
	return message.NewPrinter(tag, message.Catalog(cat)), confidence == language.Exact
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
