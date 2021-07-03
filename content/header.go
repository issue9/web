// SPDX-License-Identifier: MIT

package content

import (
	"mime"
	"strings"

	"github.com/issue9/qheader"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
)

// DefaultCharset 默认的字符集
const DefaultCharset = "utf-8"

// CharsetIsNop 指定的编码是否不需要任何额外操作
func CharsetIsNop(enc encoding.Encoding) bool {
	return enc == nil || enc == unicode.UTF8 || enc == encoding.Nop
}

// AcceptCharset 根据 Accept-Charset 报头的内容获取其最值的字符集信息
//
// 传递 * 获取返回默认的字符集相关信息，即 utf-8
// 其它值则按值查找，或是在找不到时返回空值。
//
// 返回的 name 值可能会与 header 中指定的不一样，比如 gb_2312 会被转换成 gbk
func AcceptCharset(header string) (name string, enc encoding.Encoding) {
	if header == "" || header == "*" {
		return DefaultCharset, nil
	}

	var err error
	accepts := qheader.Parse(header, "*")
	for _, apt := range accepts {
		enc, err = htmlindex.Get(apt.Value)
		if err != nil { // err != nil 表示未找到，继续查找
			continue
		}

		// 转换成官方的名称
		name, err = htmlindex.Name(enc)
		if err != nil {
			name = apt.Value // 不存在，直接使用用户上传的名称
		}

		return name, enc
	}

	return "", nil
}

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
func ParseContentType(v string) (mimetype, charset string, err error) {
	if v = strings.TrimSpace(v); v == "" {
		return DefaultMimetype, DefaultCharset, nil
	}

	mt, params, err := mime.ParseMediaType(v)
	if err != nil {
		return "", "", err
	}
	if charset = params["charset"]; charset == "" {
		charset = DefaultCharset
	}
	return mt, charset, nil
}

// BuildContentType 生成一个 content-type
//
// 若值为空，则会使用默认值代替
func BuildContentType(mt, charset string) string {
	if mt == "" {
		mt = DefaultMimetype
	}
	if charset == "" {
		charset = DefaultCharset
	}

	return mt + "; charset=" + charset
}
