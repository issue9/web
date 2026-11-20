// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package locale 用于处理本地化文件的加载
package locale

import (
	"io/fs"
	"strings"
	"time"

	"github.com/issue9/config"
	"github.com/jellydator/ttlcache/v3"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

const maxPrinters = 64 // 缓存的 [message.Printer] 数量上限

const maxPrinterIdle = 10 * time.Minute // 缓存项的空闲时长

type Locale struct {
	builder              *catalog.Builder
	id                   language.Tag
	printer              *message.Printer
	printers             *ttlcache.Cache[language.Tag, *message.Printer]
	config               *config.Config
	acceptLanguageHeader string // 根据 builder 加载的语言文件，生成服务端的 Accept-Language 报头
}

func New(lang language.Tag, conf *config.Config) *Locale {
	b := catalog.NewBuilder(catalog.Fallback(lang))

	// 保证 b 中包含一条 id 语言的翻译项，这样可以始终让 Negotiation.LocalePrinter 的对象始终是有值的。
	if err := b.SetString(lang, "_____", "_____"); err != nil {
		panic(err)
	}

	lang, _, _ = b.Matcher().Match(lang) // 从 b 中查找最合适的 tag

	return &Locale{
		builder: b,
		id:      lang,
		config:  conf,
		printer: message.NewPrinter(lang, message.Catalog(b)),
		printers: ttlcache.New(
			ttlcache.WithCapacity[language.Tag, *message.Printer](maxPrinters),
			ttlcache.WithTTL[language.Tag, *message.Printer](maxPrinterIdle),
		),
		acceptLanguageHeader: buildAcceptLanguage(b),
	}
}

func buildAcceptLanguage(ls catalog.Catalog) string {
	var s strings.Builder
	for index, v := range ls.Languages() {
		if index == 0 {
			s.WriteString(v.String())
		} else {
			s.WriteString(", ")
			s.WriteString(v.String())
		}
	}
	return s.String()
}

func (l *Locale) ID() language.Tag { return l.id }

func (l *Locale) Catalog() *catalog.Builder { return l.builder }

func (l *Locale) Config() *config.Config { return l.config }

func (l *Locale) AcceptLanguageHeader() string { return l.acceptLanguageHeader }

// LoadMessages 从 fsys 中加载符合 glob 的本地化文件
func (l *Locale) LoadMessages(glob string, fsys ...fs.FS) error {
	if err := Load(l.Config().Serializer(), l.Catalog(), glob, fsys...); err != nil {
		return err
	}

	l.acceptLanguageHeader = buildAcceptLanguage(l.Catalog())
	l.printers.DeleteAll()

	return nil
}

// Printer 最符合 [Locale.ID] 的 [message.Printer] 对象
func (l *Locale) Printer() *message.Printer { return l.printer }

func (l *Locale) Sprintf(key string, v ...any) string { return l.Printer().Sprintf(key, v...) }

// 每当更新 [Locale.Catalog] 时，可能造成 NewPrinter 相同的入参而返回不同的返回对象的情况。
func (l *Locale) NewPrinter(id language.Tag) *message.Printer {
	if id == l.ID() {
		return l.Printer()
	}

	if item := l.printers.Get(id); item != nil {
		return item.Value()
	}

	// 从 l 中查找最合适的 tag
	id, _, _ = l.Catalog().Matcher().Match(id)
	if item := l.printers.Get(id); item != nil {
		return item.Value()
	}

	p := message.NewPrinter(id, message.Catalog(l.Catalog()))
	l.printers.Set(id, p, maxPrinterIdle)
	return p
}

func (l *Locale) AcceptLanguage(h string) language.Tag {
	if h == "" {
		return l.ID()
	}
	tag, _ := language.MatchStrings(l.Catalog().Matcher(), h)
	return tag
}
