// SPDX-License-Identifier: MIT

// Package content 提供对各类媒体数据的处理
package content

import (
	"strings"
	"unicode"

	"github.com/issue9/qheader"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
)

// AcceptLanguage 从 accept-language 报头中获取最适合的本地化语言信息
func AcceptLanguage(cl catalog.Catalog, header string) language.Tag {
	if header == "" {
		return language.Und
	}

	al := qheader.Parse(header, "*")
	tags := make([]language.Tag, 0, len(al))
	for _, l := range al {
		tags = append(tags, language.Make(l.Value))
	}

	tag, _, _ := cl.Matcher().Match(tags...)
	return tag
}

// ParseContentType 从 content-type 中获取编码和字符集
//
// 若客户端传回的是空值，则会使用默认值代替。
//
// 返回值中，mimetype 一律返回小写的值，charset 则原样返回
//
// https://tools.ietf.org/html/rfc7231#section-3.1.1.1
func ParseContentType(v string) (mime, charset string, err error) {
	if v = strings.ToLower(strings.TrimSpace(v)); v == "" {
		return DefaultMimetype, DefaultCharset, nil
	}

	index := strings.IndexByte(v, ';')
	switch {
	case index < 0: // 只有编码
		return strings.ToLower(v), DefaultCharset, nil
	case index == 0: // mimetype 不可省略
		return "", "", errContentTypeMissMimetype
	}

	mime = strings.ToLower(v[:index])

	for index > 0 {
		// 去掉左边的空白字符
		v = strings.TrimLeftFunc(v[index+1:], func(r rune) bool { return unicode.IsSpace(r) })

		if !strings.HasPrefix(v, "charset=") { // 按规定，不用考虑 = 两边没有空白字符。
			index = strings.IndexByte(v, ';')
			continue
		}

		v = strings.TrimPrefix(v, "charset=")
		return mime, strings.TrimFunc(v, func(r rune) bool { return r == '"' }), nil
	}

	return mime, DefaultCharset, nil
}

// BuildContentType 生成一个 content-type
//
// 若值为空，则会使用默认值代替
func BuildContentType(mime, charset string) string {
	if mime == "" {
		mime = DefaultMimetype
	}
	if charset == "" {
		charset = DefaultCharset
	}

	return mime + "; charset=" + charset
}
