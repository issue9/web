// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"errors"

	"github.com/issue9/middleware/compress/accept"
	xencoding "golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/encoding"
)

var errInvalidCharset = errors.New("无效的字符集")

// 指定的编码是否不需要任何额外操作
func charsetIsNop(enc xencoding.Encoding) bool {
	return enc == nil ||
		enc == unicode.UTF8 ||
		enc == xencoding.Nop
}

// 根据 Accept-Charset 报头的内容获取其最值的字符集信息。
//
// 传递 * 获取返回默认的字符集相关信息，即 utf-8
// 其它值则按值查找，或是在找不到时返回空值。
//
// 返回的 name 值可能会与 header 中指定的不一样，比如 gb_2312 会被转换成 gbk
func acceptCharset(header string) (name string, enc xencoding.Encoding, err error) {
	if header == "" || header == "*" {
		return encoding.DefaultCharset, nil, nil
	}

	accepts, err := accept.Parse(header)
	if err != nil {
		return "", nil, err
	}

	for _, accept := range accepts {
		enc, err = htmlindex.Get(accept.Value)
		if err != nil { // err != nil 表示未找到，继续查找
			continue
		}

		// 转换成官方的名称
		name, err = htmlindex.Name(enc)
		if err != nil {
			name = accept.Value // 不存在，直接使用用户上传的名称
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
