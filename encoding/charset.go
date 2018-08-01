// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"strings"

	"github.com/issue9/web/internal/accept"
	xencoding "golang.org/x/text/encoding"
)

// DefaultCharset 默认的字符集，在不能正确获取输入和输出的字符集时，
// 会采用此值和为其默认值。
const DefaultCharset = "UTF-8"

// Charseter 字符集编码需要实现的接口
type Charseter = xencoding.Encoding

var charset = make(map[string]Charseter, 10)

func init() {
	if err := AddCharset(DefaultCharset, xencoding.Nop); err != nil {
		panic(err)
	}
}

func findCharset(name string) (string, Charseter) {
	if name == "*" {
		return DefaultCharset, charset[DefaultCharset]
	}

	name = strings.ToUpper(name)
	return name, charset[name]
}

// AddCharset 添加字符集
func AddCharset(name string, c Charseter) error {
	if name == "*" {
		return ErrInvalidCharset
	}

	name = strings.ToUpper(name)
	if _, found := charset[name]; found {
		return ErrExists
	}

	charset[name] = c

	return nil
}

// AddCharsets 添加多个字符集
func AddCharsets(cs map[string]Charseter) error {
	for k, v := range cs {
		if err := AddCharset(k, v); err != nil {
			return err
		}
	}

	return nil
}

// AcceptCharset 根据 Accept-Charset 报头的内容获取其最值的字符集信息。
//
// 传递 * 获取返回默认的字符集相关信息，即 DefaultCharset
// 其它值则按值查找，或是在找不到时返回空值。
func AcceptCharset(header string) (name string, enc Charseter, err error) {
	if header == "" || header == "*" {
		name, enc := findCharset("*")
		return name, enc, nil
	}

	accepts, err := accept.Parse(header)
	if err != nil {
		return "", nil, err
	}

	for _, accept := range accepts {
		name, enc := findCharset(accept.Value)
		if enc != nil {
			return name, enc, nil
		}
	}

	return "", nil, ErrInvalidCharset
}
