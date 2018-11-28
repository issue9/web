// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package encoding 提供了框架内对编码和字符集功能的支持。
package encoding

import (
	"errors"
	"strings"
	"unicode"
)

// DefaultCharset 默认的字符集
const DefaultCharset = "utf-8"

// Nil 表示向客户端输出 nil 值。
//
// 这是一个只有类型但是值为空的变量。在某些特殊情况下，
// 如果需要向客户端输出一个 nil 值的内容，可以使用此值。
var Nil *struct{}

var (
	// ErrExists 表示指定名称的项目已经存在。
	//
	// 在 AddCharset、Addmarshal 和 AddUnmarshal 中会返回此错误。
	ErrExists = errors.New("该名称的项目已经存在")

	// ErrInvalidMimeType 无效的 mimetype 值，一般为 content-type 或
	// Accept 等报头指定的 mimetype 值无效。
	ErrInvalidMimeType = errors.New("mimetype 无效")
)

// Unmarshal 查找指定名称的 UnmarshalFunc
func Unmarshal(name string) (UnmarshalFunc, error) {
	var unmarshal *unmarshaler
	for _, mt := range unmarshals {
		if mt.name == name {
			unmarshal = mt
			break
		}
	}
	if unmarshal == nil {
		return nil, ErrInvalidMimeType
	}

	return unmarshal.f, nil
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

	if v == "" {
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
