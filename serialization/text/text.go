// SPDX-License-Identifier: MIT

// Package text 针对文本内容的编解码实现
package text

import (
	"encoding"

	"github.com/issue9/web/serialization"
)

// Mimetype 当前包能解析的编码类型
const Mimetype = "text/plain"

// Marshal 针对文本内容的 MarshalFunc 实现
func Marshal(v interface{}) ([]byte, error) {
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

	return nil, serialization.ErrUnsupported
}

// Unmarshal 针对文本内容的 UnmarshalFunc 实现
func Unmarshal(data []byte, v interface{}) (err error) {
	switch vv := v.(type) {
	case *string:
		*vv = string(data)
	case *[]byte:
		*vv = data
	case *[]rune:
		*vv = []rune(string(data))
	case encoding.TextUnmarshaler:
		err = vv.UnmarshalText(data)
	default:
		err = serialization.ErrUnsupported
	}

	return err
}
