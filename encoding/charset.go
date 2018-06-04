// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"github.com/issue9/web/internal/accept"
	xencoding "golang.org/x/text/encoding"
)

// DefaultCharset 默认的字符集，在不能正确获取输入和输出的字符集时，
// 会采用此值和为其默认值。
const DefaultCharset = "utf-8"

var charset = map[string]xencoding.Encoding{
	DefaultCharset: xencoding.Nop,
}

// AddCharset 添加字符集
func AddCharset(name string, c xencoding.Encoding) error {
	if _, found := charset[name]; found {
		return ErrExists
	}

	charset[name] = c

	return nil
}

// AcceptCharset 根据 Accept-Charset 报头的内容获取其最值的字符集信息。
func AcceptCharset(header string) (name string, enc xencoding.Encoding, err error) {
	if header == "" {
		header = DefaultCharset
	}

	accepts, err := accept.Parse(header)
	if err != nil {
		return "", nil, err
	}

	for _, accept := range accepts {
		if enc := charset[accept.Value]; enc != nil {
			return accept.Value, enc, nil
		}
	}

	return "", nil, ErrUnsupportedCharset
}
