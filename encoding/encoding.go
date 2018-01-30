// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package encoding 提供了框架内对编码和字符集功能的支持。
package encoding

import (
	"errors"
	"strings"
	"unicode"

	"golang.org/x/text/encoding"
)

const (
	// DefaultMimeType 默认的媒体类型，在不能正确获取输入和输出的媒体类型时，
	// 会采用此值作为其默认值。
	DefaultMimeType = "text/plain"

	// DefaultCharset 默认的字符集，在不能正确获取输入和输出的字符集时，
	// 会采用此值和为其默认值。
	DefaultCharset = "utf-8"
)

var (
	// ErrExists 表示指定名称的项目已经存在
	// 在 AddCharset、Addmarshal 和 AddUnmarshal 中会返回此错误。
	ErrExists = errors.New("该名称的项目已经存在")

	// ErrUnsupportedMarshal MarshalFunc 和 UnmarshalFunc 的实现者中，
	// 如果无法识别数据内容，则返回此错误信息。
	ErrUnsupportedMarshal = errors.New("对象没有有效的转换方法")
)

var (
	charset = map[string]encoding.Encoding{
		DefaultCharset: encoding.Nop,
	}

	marshals = map[string]MarshalFunc{
		DefaultMimeType: TextMarshal,
	}

	unmarshals = map[string]UnmarshalFunc{
		DefaultMimeType: TextUnmarshal,
	}
)

// MarshalFunc 将一个对象转换成 []byte 内容时，所采用的接口。
type MarshalFunc func(v interface{}) ([]byte, error)

// UnmarshalFunc 将客户端内容转换成一个对象时，所采用的接口。
type UnmarshalFunc func([]byte, interface{}) error

// Charset 获取指定名称的字符集
// 若不存在，则返回 nil
func Charset(name string) encoding.Encoding {
	return charset[name]
}

// AddCharset 添加字符集
func AddCharset(name string, c encoding.Encoding) error {
	if _, found := charset[name]; found {
		return ErrExists
	}

	charset[name] = c

	return nil
}

// Marshal 获取指定名称的编码函数
func Marshal(name string) MarshalFunc {
	return marshals[name]
}

// AddMarshal 添加编码函数
func AddMarshal(name string, m MarshalFunc) error {
	if _, found := marshals[name]; found {
		return ErrExists
	}

	marshals[name] = m
	return nil
}

// Unmarshal 获取指定名称的编码函数
func Unmarshal(name string) UnmarshalFunc {
	return unmarshals[name]
}

// AddUnmarshal 添加编码函数
func AddUnmarshal(name string, m UnmarshalFunc) error {
	if _, found := unmarshals[name]; found {
		return ErrExists
	}

	unmarshals[name] = m
	return nil
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
		return "", "", errors.New("缺少 mimetype")
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
