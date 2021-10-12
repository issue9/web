// SPDX-License-Identifier: MIT

// Package content 与生成内容相关的功能
package content

import (
	"time"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/serialization"
)

// DefaultMimetype 默认的媒体类型
//
// 在不能获取输入和输出的媒体类型时，会采用此值作为其默认值。
const DefaultMimetype = "application/octet-stream"

// Content 管理反馈给用户的数据
type Content struct {
	mimetypes *serialization.Mimetypes

	// 本地化相关信息
	location      *time.Location
	locale        *serialization.Locale
	tag           language.Tag
	localePrinter *message.Printer

	resultMessages map[int]*resultMessage
	resultBuilder  BuildResultFunc
}

// New 返回 *Content 实例
//
// files 加载本地化数据的方法；
// tag 默认的本地化语言标签，context 查找不到数据时采用此值，同时也作为 context 实例的默认输出语言。
func New(builder BuildResultFunc, loc *time.Location, files *serialization.Files, tag language.Tag) *Content {
	b := catalog.NewBuilder(catalog.Fallback(tag))

	return &Content{
		mimetypes: serialization.NewMimetypes(10),

		location:      loc,
		locale:        serialization.NewLocale(b, files),
		tag:           tag,
		localePrinter: message.NewPrinter(tag, message.Catalog(b)),

		resultMessages: make(map[int]*resultMessage, 20),
		resultBuilder:  builder,
	}
}

// Mimetypes 管理 mimetype 的序列化操作
func (c *Content) Mimetypes() *serialization.Mimetypes { return c.mimetypes }

// Files 返回用于序列化文件内容的操作接口
func (c *Content) Files() *serialization.Files { return c.Locale().Files() }

func (c *Content) Location() *time.Location { return c.location }

// 指定的编码是否不需要任何额外操作
func charsetIsNop(enc encoding.Encoding) bool {
	return enc == nil || enc == unicode.UTF8 || enc == encoding.Nop
}

// 生成 content-type，若值为空，则会使用默认值代替。
func buildContentType(mt, charset string) string {
	if mt == "" {
		mt = DefaultMimetype
	}
	if charset == "" {
		charset = DefaultCharset
	}

	return mt + "; charset=" + charset
}
