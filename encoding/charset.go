// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"github.com/issue9/web/internal/accept"
	xencoding "golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/unicode"
)

// defaultCharset 默认的字符集，在不能正确获取输入和输出的字符集时，
// 会采用此值和为其默认值。
const defaultCharset = "utf-8"

// CharsetIsNop 指定的编码是否不需要任何额外操作
func CharsetIsNop(enc xencoding.Encoding) bool {
	return enc == nil ||
		enc == unicode.UTF8 ||
		enc == xencoding.Nop
}

// AddCharset 添加字符集
// Deprecated: 不再启作用
func AddCharset(name string, c xencoding.Encoding) error {
	return nil
}

// AddCharsets 添加多个字符集
// Deprecated: 不再启作用
func AddCharsets(cs map[string]xencoding.Encoding) error {
	return nil
}

// AcceptCharset 根据 Accept-Charset 报头的内容获取其最值的字符集信息。
//
// 传递 * 获取返回默认的字符集相关信息，即 defaultCharset
// 其它值则按值查找，或是在找不到时返回空值。
//
// 返回的 name 值可能会与 header 中指定的不一样，比如 gb_2312 会被转换成 gbk
func AcceptCharset(header string) (name string, enc xencoding.Encoding, err error) {
	if header == "" || header == "*" {
		return defaultCharset, nil, nil
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

	return "", nil, ErrInvalidCharset
}
