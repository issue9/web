// SPDX-License-Identifier: MIT

// Package gob 提供 GOB 格式的编解码
package gob

import (
	"bytes"
	"encoding/gob"
)

const Mimetype = "application/octet-stream"

func Marshal(v interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	enc := gob.NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func Unmarshal(data []byte, v interface{}) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}
