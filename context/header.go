// SPDX-License-Identifier: MIT

package context

import (
	"errors"
	"strings"
	"unicode"

	"github.com/issue9/middleware/compress/accept"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	xunicode "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/context/mimetype"
)

const utfName = "utf-8"

var errInvalidCharset = errors.New("无效的字符集")

// 指定的编码是否不需要任何额外操作
func charsetIsNop(enc encoding.Encoding) bool {
	return enc == nil ||
		enc == xunicode.UTF8 ||
		enc == encoding.Nop
}

// 根据 Accept-Charset 报头的内容获取其最值的字符集信息。
//
// 传递 * 获取返回默认的字符集相关信息，即 utf-8
// 其它值则按值查找，或是在找不到时返回空值。
//
// 返回的 name 值可能会与 header 中指定的不一样，比如 gb_2312 会被转换成 gbk
func acceptCharset(header string) (name string, enc encoding.Encoding, err error) {
	if header == "" || header == "*" {
		return utfName, nil, nil
	}

	accepts, err := accept.Parse(header)
	if err != nil {
		return "", nil, err
	}

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

		return name, enc, nil
	}

	return "", nil, errInvalidCharset
}

func acceptLanguage(header string) (language.Tag, error) {
	if header == "" {
		return language.Und, nil
	}

	al, err := accept.Parse(header)
	if err != nil {
		return language.Und, err
	}

	prefs := make([]language.Tag, 0, len(al))
	for _, l := range al {
		prefs = append(prefs, language.Make(l.Value))
	}

	tag, _, _ := message.DefaultCatalog.Matcher().Match(prefs...)
	return tag, nil
}

var errContentTypeMissMimetype = errors.New("content-type 不存在 mimetype 部分")

// 从 content-type 中获取编码和字符集
//
// 若客户端传回的是空值，则会使用默认值代替。
//
// 返回值中，mimetype 一律返回小写的值，charset 则原样返回
//
// https://tools.ietf.org/html/rfc7231#section-3.1.1.1
func parseContentType(v string) (mime, charset string, err error) {
	v = strings.TrimSpace(v)

	if v == "" {
		return mimetype.DefaultMimetype, utfName, nil
	}

	index := strings.IndexByte(v, ';')
	switch {
	case index < 0: // 只有编码
		return strings.ToLower(v), utfName, nil
	case index == 0: // mimetype 不可省略
		return "", "", errContentTypeMissMimetype
	}

	mime = strings.ToLower(v[:index])

	for index > 0 {
		// 去掉左边的空白字符
		v = strings.TrimLeftFunc(v[index+1:], func(r rune) bool { return unicode.IsSpace(r) })

		if !strings.HasPrefix(v, "charset=") {
			index = strings.IndexByte(v, ';')
			continue
		}

		v = strings.TrimPrefix(v, "charset=")
		return mime, strings.TrimFunc(v, func(r rune) bool { return r == '"' }), nil
	}

	return mime, utfName, nil
}

// 生成一个 content-type
//
// 若值为空，则会使用默认值代替
func buildContentType(mime, charset string) string {
	if mime == "" {
		mime = mimetype.DefaultMimetype
	}
	if charset == "" {
		charset = utfName
	}

	return mime + "; charset=" + charset
}
