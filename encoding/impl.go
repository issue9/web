// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import stdencoding "encoding"

// TextMarshal 针对文本内容的 MarshalFunc 实现
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

// TextUnmarshal 针对文本内容的 UnmarshalFunc 实现
func TextUnmarshal(data []byte, v interface{}) error {
	if vv, ok := v.(stdencoding.TextUnmarshaler); ok {
		return vv.UnmarshalText(data)
	}

	return ErrUnsupportedMarshal
}
