// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package gob 提供 GOB 格式的编解码
package gob

import (
	"bytes"
	"encoding/gob"
)

// MimeType 当前编码默认情况下使用的编码名称。
const MimeType = "application/octet-stream"

// Marshal 针对 GOB 内容的 MarshalFunc 实现
func Marshal(v interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	enc := gob.NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// Unmarshal 针对 GOB 内容的 UnmarshalFunc 实现
func Unmarshal(data []byte, v interface{}) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}
