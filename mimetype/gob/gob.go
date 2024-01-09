// SPDX-License-Identifier: MIT

// Package gob [GOB] 格式的数据编码方案
//
// [GOB]: https://pkg.go.dev/encoding/gob
package gob

import (
	"bytes"
	"encoding/gob"
	"io"

	"github.com/issue9/web"
)

const Mimetype = "application/octet-stream"

func Marshal(_ *web.Context, v any) ([]byte, error) {
	w := new(bytes.Buffer)
	enc := gob.NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func Unmarshal(r io.Reader, v any) error { return gob.NewDecoder(r).Decode(v) }
