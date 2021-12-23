// SPDX-License-Identifier: MIT

// Package text 针对文本内容的编解码实现
package text

import (
	"encoding"

	"github.com/issue9/web/serialization"
)

const Mimetype = "text/plain"

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
	case error:
		return nil, vv
	}

	return nil, serialization.ErrUnsupported
}

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
