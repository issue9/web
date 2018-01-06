// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"errors"
	"strings"

	"golang.org/x/text/encoding"
)

const (
	// DefaultEncoding 默认的编码方式，在不能正确获取输入和输出的编码方式时，
	// 会采用此值作为其默认值。
	DefaultEncoding = "application/json"

	// DefaultCharset 默认的字符集，在不能正确获取输入和输出的字符集时，
	// 会采用此值和为其默认值。
	DefaultCharset = "utf-8"
)

// Marshal 将一个对象转换成 []byte 内容时，所采用的接口。
// Context.Render() 会调用此接口。
type Marshal func(v interface{}) ([]byte, error)

// Unmarshal 将客户端内容转换成一个对象时，所采用的接口。
// Context.Read() 会调用此接口。
type Unmarshal func([]byte, interface{}) error

// ErrExists 表示存在相同名称的项。
// 多个类似功能的函数，都有可能返回此错误。
var ErrExists = errors.New("存在相同名称的项")

var (
	marshals   = map[string]Marshal{}
	unmarshals = map[string]Unmarshal{}

	charset = map[string]encoding.Encoding{
		DefaultCharset: nil,
	}
)

// AddMarshal 添加一个新的解码器，只有通过 AddMarshal 添加的解码器，
// 才能被 Context 使用。
func AddMarshal(name string, m Marshal) error {
	_, found := marshals[name]
	if found {
		return ErrExists
	}

	marshals[name] = m
	return nil
}

// AddUnmarshal 添加一个编码器，只有通过 AddUnmarshal 添加的解码器，
// 才能被 Context 使用。
func AddUnmarshal(name string, m Unmarshal) error {
	_, found := unmarshals[name]
	if found {
		return ErrExists
	}

	unmarshals[name] = m
	return nil
}

// AddCharset 添加编码方式，只有通过 AddCharset 添加的字符集，
// 才能被 Context 使用。
func AddCharset(name string, enc encoding.Encoding) error {
	_, found := charset[name]
	if found {
		return ErrExists
	}

	charset[name] = enc
	return nil
}

// 生成一个 content-type
func buildContentType(encoding, charset string) string {
	if encoding == "" {
		encoding = DefaultEncoding
	}
	if charset == "" {
		charset = DefaultCharset
	}

	return encoding + ";charset=" + charset
}

// 从 content-type 中获取编码和字符集
func parseContentType(v string) (encoding, charset string) {
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
