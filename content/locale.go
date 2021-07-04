// SPDX-License-Identifier: MIT

package content

import (
	"io"

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

// Fprint 相当于 ctx.LocalePrinter.Fprint
func (ctx *Context) Fprint(w io.Writer, v ...interface{}) (int, error) {
	return ctx.LocalePrinter.Fprint(w, v...)
}

// Fprintf 相当于 ctx.LocalePrinter.Fprintf
func (ctx *Context) Fprintf(w io.Writer, key message.Reference, v ...interface{}) (int, error) {
	return ctx.LocalePrinter.Fprintf(w, key, v...)
}

// Fprintln 相当于 ctx.LocalePrinter.Fprintln
func (ctx *Context) Fprintln(w io.Writer, v ...interface{}) (int, error) {
	return ctx.LocalePrinter.Fprintln(w, v...)
}

// Print 相当于 ctx.LocalePrinter.Print
func (ctx *Context) Print(v ...interface{}) (int, error) {
	return ctx.LocalePrinter.Print(v...)
}

// Printf 相当于 ctx.LocalePrinter.Printf
func (ctx *Context) Printf(key message.Reference, v ...interface{}) (int, error) {
	return ctx.LocalePrinter.Printf(key, v...)
}

// Println 相当于 ctx.LocalePrinter.Println
func (ctx *Context) Println(v ...interface{}) (int, error) {
	return ctx.LocalePrinter.Println(v...)
}

// Sprint 相当于 ctx.LocalePrinter.Sprint
func (ctx *Context) Sprint(v ...interface{}) string {
	return ctx.LocalePrinter.Sprint(v...)
}

// Sprintf 相当于 ctx.LocalePrinter.Sprintf
func (ctx *Context) Sprintf(key message.Reference, v ...interface{}) string {
	return ctx.LocalePrinter.Sprintf(key, v...)
}

// Sprintln 相当于 ctx.LocalePrinter.Sprintln
func (ctx *Context) Sprintln(v ...interface{}) string {
	return ctx.LocalePrinter.Sprintln(v...)
}
