// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"io/fs"
	"strings"
	"time"

	"github.com/issue9/config"
	"github.com/jellydator/ttlcache/v3"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/locale"
)

// maxPrinters 缓存的 [message.Printer] 数量上限
const maxPrinters = 64

// maxPrinterIdle 缓存项的空闲时长
const maxPrinterIdle = 10 * time.Minute

// Locale 提供与本地化相关的功能
type Locale struct {
	*catalog.Builder
	id       language.Tag
	s        config.Serializer
	printer  *message.Printer
	printers *ttlcache.Cache[language.Tag, *message.Printer]

	acceptLanguage string // 根据内容生成 Accept-Language 报头
}

func newLocale(id language.Tag, s config.Serializer) *Locale {
	b := catalog.NewBuilder(catalog.Fallback(id))

	// 保证 b 中包含一条 id 语言的翻译项，
	// 这样可以始终让 Locale.Printer 的对象始终是有值的。
	if err := b.SetString(id, "_____", "_____"); err != nil {
		panic(err)
	}

	id, _, _ = b.Matcher().Match(id) // 从 b 中查找最合适的 tag
	return &Locale{
		Builder: b,
		id:      id,
		s:       s,
		printer: message.NewPrinter(id, message.Catalog(b)),
		printers: ttlcache.New(
			ttlcache.WithCapacity[language.Tag, *message.Printer](maxPrinters),
			ttlcache.WithTTL[language.Tag, *message.Printer](maxPrinterIdle),
		),
		acceptLanguage: buildAcceptLanguage(b),
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

// ID 返回默认的语言标签
func (l *Locale) ID() language.Tag { return l.id }

// LoadMessages 从 fsys 中加载符合 glob 的本地化文件
//
// 根据 [Server.Config] 处理文件格式，如果文件格式不被 [Server.Config] 支持，将无法加载。
func (l *Locale) LoadMessages(glob string, fsys ...fs.FS) error {
	if err := locale.Load(l.s, l.Builder, glob, fsys...); err != nil {
		return err
	}

	l.acceptLanguage = buildAcceptLanguage(l.Builder)
	l.printers.DeleteAll()

	return nil
}

// Printer 最符合 [Locale.ID] 的 [message.Printer] 对象
func (l *Locale) Printer() *message.Printer { return l.printer }

// Sprintf 等同于 [Locale.Printer.Sprintf]
func (l *Locale) Sprintf(key string, v ...any) string { return l.Printer().Sprintf(key, v...) }

// NewPrinter 声明最符合 tag 的 [message.Printer] 对象
//
// NOTE: 每当给 [Locale.SetString]、[Locale.SetMacro] 和 [Locale.Set] 传递新的 [language.Tag]
// 值时，可能造成 NewPrinter 相同的入参而返回不同的返回对象的情况。
func (l *Locale) NewPrinter(id language.Tag) *message.Printer {
	if id == l.ID() {
		return l.Printer()
	}

	if item := l.printers.Get(id); item != nil {
		return item.Value()
	}

	// 从 l 中查找最合适的 tag
	id, _, _ = l.Matcher().Match(id)
	if item := l.printers.Get(id); item != nil {
		return item.Value()
	}

	p := message.NewPrinter(id, message.Catalog(l))
	l.printers.Set(id, p, maxPrinterIdle)
	return p
}

// SetString 添加新的翻译项
func (l *Locale) SetString(tag language.Tag, key, msg string) error {
	if err := l.Builder.SetString(tag, key, msg); err != nil {
		return err
	}
	l.printers.DeleteAll()
	return nil
}

// SetMacro 添加新的翻译项
func (l *Locale) SetMacro(tag language.Tag, name string, msg ...catalog.Message) error {
	if err := l.Builder.SetMacro(tag, name, msg...); err != nil {
		return err
	}
	l.printers.DeleteAll()
	return nil
}

// Set 添加新的翻译项
func (l *Locale) Set(tag language.Tag, key string, msg ...catalog.Message) error {
	if err := l.Builder.Set(tag, key, msg...); err != nil {
		return err
	}
	l.printers.DeleteAll()
	return nil
}
