// SPDX-License-Identifier: MIT

// Package gob 提供 GOB 格式的编解码
package gob

import (
	"bytes"
	"encoding/gob"
)

// Mimetype 当前编码默认情况下使用的编码名称
const Mimetype = "application/octet-stream"

// Marshal 针对 GOB 内容的 content.MarshalFunc 实现
func Marshal(v interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	enc := gob.NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// Unmarshal 针对 GOB 内容的 content.UnmarshalFunc 实现
func Unmarshal(data []byte, v interface{}) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}
