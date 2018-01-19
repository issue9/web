// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package encoding 提供了框架内对编码和字符集功能的支持。
package encoding

import (
	stdencoding "encoding"
	"errors"
	"strings"

	"golang.org/x/text/encoding"
)

// Marshal 和 Unmarshal 的实现者中，如果无法识别数据内容，
// 则返回此错误信息。
var ErrUnsupportedMarshal = errors.New("对象没有有效的转换方法")

const (
	// DefaultEncoding 默认的编码方式，在不能正确获取输入和输出的编码方式时，
	// 会采用此值作为其默认值。
	DefaultEncoding = "text/plain"

	// DefaultCharset 默认的字符集，在不能正确获取输入和输出的字符集时，
	// 会采用此值和为其默认值。
	DefaultCharset = "utf-8"
)

// Marshal 将一个对象转换成 []byte 内容时，所采用的接口。
type Marshal func(v interface{}) ([]byte, error)

// Unmarshal 将客户端内容转换成一个对象时，所采用的接口。
type Unmarshal func([]byte, interface{}) error

// Charset 操作字符集的相关内容
//
// golang.org/x/text 下包含部分常用的结构。
type Charset = encoding.Encoding

// TextMarshal 针对文本内容的 Marshal 实现
func TextMarshal(v interface{}) ([]byte, error) {
	switch vv := v.(type) {
	case string:
		return []byte(vv), nil
	case []byte:
		return vv, nil
	case []rune:
		return []byte(string(vv)), nil
	case stdencoding.TextMarshaler:
		return vv.MarshalText()
	}

	return nil, ErrUnsupportedMarshal
}

// TextUnmarshal 针对文本内容的 Marshal 实现
func TextUnmarshal(data []byte, v interface{}) error {
	if vv, ok := v.(stdencoding.TextUnmarshaler); ok {
		return vv.UnmarshalText(data)
	}

	return ErrUnsupportedMarshal
}

// BuildContentType 生成一个 content-type
//
// 若值为空，则会使用默认值代替
func BuildContentType(encoding, charset string) string {
	if encoding == "" {
		encoding = DefaultEncoding
	}
	if charset == "" {
		charset = DefaultCharset
	}

	return encoding + "; charset=" + charset
}

// ParseContentType 从 content-type 中获取编码和字符集
//
// 若客户端传回的是空值，则会使用默认值代替。
func ParseContentType(v string) (encoding, charset string) {
	v = strings.ToLower(strings.TrimSpace(v))
	if len(v) == 0 {
		return DefaultEncoding, DefaultCharset
	}

	// encoding
	index := strings.IndexByte(v, ';')
	switch {
	case index < 0: // 只有编码
		return v, DefaultCharset
	case index == 0: // 编码为空
		encoding = DefaultEncoding
	case index > 0:
		encoding = strings.TrimSpace(v[:index])
	}

	v = v[index+1:]
	if len(v) == 0 {
		return encoding, DefaultCharset
	}

	index = strings.IndexByte(v, ';') // 查找第二个 ;
	switch {
	case index == 0:
		return encoding, DefaultCharset
	case index > 0:
		v = v[:index]
	}

	index = strings.IndexByte(v, '=')
	switch {
	case index < 0:
		charset = strings.TrimSpace(v)
	case index >= 0:
		charset = strings.TrimSpace(v[index+1:])
	}
	if charset == "" {
		charset = DefaultCharset
	}

	return encoding, charset
}
