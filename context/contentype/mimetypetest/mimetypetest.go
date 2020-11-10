// SPDX-License-Identifier: MIT

// Package mimetypetest 针对文本内容的编解码实现，仅作为测试用例。
package mimetypetest

import (
	"encoding"
	"errors"

	"github.com/issue9/web/context/contentype"
)

// Mimetype 当前包能解析的编码类型
const Mimetype = "text/plain"

var errUnsupported = errors.New("对象没有有效的转换方法")

// Nil contentype.Nil 解码后的值
var Nil = []byte("NIL")

// TextMarshal 针对文本内容的 MarshalFunc 实现
func TextMarshal(v interface{}) ([]byte, error) {
	if v == contentype.Nil {
		return Nil, nil
	}

	switch vv := v.(type) {
	case string:
		return []byte(vv), nil
	case []byte:
		return vv, nil
	case []rune:
		return []byte(string(vv)), nil
	case encoding.TextMarshaler:
		return vv.MarshalText()
	}

	return nil, errUnsupported
}

// TextUnmarshal 针对文本内容的 UnmarshalFunc 实现
func TextUnmarshal(data []byte, v interface{}) error {
	if vv, ok := v.(encoding.TextUnmarshaler); ok {
		return vv.UnmarshalText(data)
	}

	return errUnsupported
}
