// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package encodingtest 针对文本内容的编解码实现，仅作为测试用例。
package encodingtest

import (
	"encoding"
	"errors"
)

var errUnsupported = errors.New("对象没有有效的转换方法")

// TextMarshal 针对文本内容的 MarshalFunc 实现
func TextMarshal(v interface{}) ([]byte, error) {
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
