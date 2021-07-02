// SPDX-License-Identifier: MIT

package content

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

// LocaleBuilder 带 tag 值的 catalog.Builder
type LocaleBuilder struct {
	builder *catalog.Builder
	tag     language.Tag
}

func (b *LocaleBuilder) SetString(key, msg string) error {
	return b.builder.SetString(b.tag, key, msg)
}

func (b *LocaleBuilder) Set(key string, msg ...catalog.Message) error {
	return b.builder.Set(b.tag, key, msg...)
}

func (b *LocaleBuilder) SetMacro(key string, msg ...catalog.Message) error {
	return b.builder.SetMacro(b.tag, key, msg...)
}

// LocaleBuilder 声明 LocaleBuilder
func (c *Content) LocaleBuilder(tag language.Tag) *LocaleBuilder {
	return &LocaleBuilder{
		builder: c.CatalogBuilder(),
		tag:     tag,
	}
}

// CatalogBuilder 返回本地化操作的相关接口
func (c *Content) CatalogBuilder() *catalog.Builder { return c.catalog }

// NewLocalePrinter 返回指定语言的 message.Printer
func (c *Content) NewLocalePrinter(tag language.Tag) *message.Printer {
	return message.NewPrinter(tag, message.Catalog(c.CatalogBuilder()))
}
