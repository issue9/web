// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package encoding 提供了框架内对编码和字符集功能的支持。
package encoding

import (
	"errors"
	"strings"
	"unicode"

	xencoding "golang.org/x/text/encoding"
)

var (
	// ErrExists 表示指定名称的项目已经存在。
	//
	// 在 AddCharset、Addmarshal 和 AddUnmarshal 中会返回此错误。
	ErrExists = errors.New("该名称的项目已经存在")

	// ErrInvalidCharset 无效的字符集。
	//
	// 一般在 Accept-Charset 或是 Content-Type
	// 等报头中指定的字符集无效或是不被支持。
	ErrInvalidCharset = errors.New("无效的字符集")

	// ErrInvalidMimeType 无效的 mimetype 值，一般为 content-type 或
	// Accept 等报头指定的 mimetype 值无效。
	ErrInvalidMimeType = errors.New("mimetype 无效")
)

// ContentType 从 content-type 报头中解析出其使用的编码和字符集函数。
func ContentType(header string) (UnmarshalFunc, xencoding.Encoding, error) {
	encName, charsetName, err := ParseContentType(header)
	if err != nil {
		return nil, nil, err
	}

	unmarshal := findUnmarshal(encName)
	if unmarshal == nil {
		return nil, nil, ErrInvalidMimeType
	}

	_, c := findCharset(charsetName)
	if c == nil {
		return nil, nil, ErrInvalidCharset
	}

	return unmarshal.f, c, nil
}

// BuildContentType 生成一个 content-type
//
// 若值为空，则会使用默认值代替
func BuildContentType(mimetype, charset string) string {
	if mimetype == "" {
		mimetype = DefaultMimeType
	}
	if charset == "" {
		charset = DefaultCharset
	}

	return mimetype + "; charset=" + charset
}

// ParseContentType 从 content-type 中获取编码和字符集
//
// 若客户端传回的是空值，则会使用默认值代替。
//
// 返回值中，mimetype 一律返回小写的值，charset 则原样返回
//
// https://tools.ietf.org/html/rfc7231#section-3.1.1.1
func ParseContentType(v string) (mimetype, charset string, err error) {
	v = strings.TrimSpace(v)

	if len(v) == 0 {
		return DefaultMimeType, DefaultCharset, nil
	}

	index := strings.IndexByte(v, ';')
	switch {
	case index < 0: // 只有编码
		return strings.ToLower(v), DefaultCharset, nil
	case index == 0: // mimetype 不可省略
		return "", "", ErrInvalidMimeType
	}

	mimetype = strings.ToLower(v[:index])

	for index > 0 {
		// 去掉左边的空白字符
		v = strings.TrimLeftFunc(v[index+1:], func(r rune) bool { return unicode.IsSpace(r) })

		if !strings.HasPrefix(v, "charset=") {
			index = strings.IndexByte(v, ';')
			continue
		}

		v = strings.TrimPrefix(v, "charset=")
		return mimetype, strings.TrimFunc(v, func(r rune) bool { return r == '"' }), nil
	}

	return mimetype, DefaultCharset, nil
}
