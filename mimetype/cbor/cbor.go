// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package cbor [CBOR] 编码
//
// [CBOR]: https://www.rfc-editor.org/rfc/rfc8949.html
package cbor

import (
	"io"

	"github.com/fxamacker/cbor/v2"

	"github.com/issue9/web"
)

const (
	Mimetype        = "application/cbor"
	ProblemMimetype = "application/problem+cbor"
)

func Marshal(_ *web.Context, v any) ([]byte, error) { return cbor.Marshal(v) }

func Unmarshal(r io.Reader, v any) error { return cbor.NewDecoder(r).Decode(v) }
