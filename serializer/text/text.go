// SPDX-License-Identifier: MIT

// Package text 针对文本内容的编解码实现
package text

import (
	"encoding"

	"github.com/issue9/web/serializer"
)

const Mimetype = "text/plain"

func Marshal(v any) ([]byte, error) {
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

	return nil, serializer.ErrUnsupported
}

func Unmarshal(data []byte, v any) (err error) {
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
		err = serializer.ErrUnsupported
	}

	return err
}