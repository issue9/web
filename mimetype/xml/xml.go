// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package xml XML 编码的序列化操作
package xml

import (
	"encoding/xml"
	"io"

	"github.com/issue9/web"
)

const (
	Mimetype        = "application/xml"
	ProblemMimetype = "application/problem+xml"
)

func Marshal(_ *web.Context, v any) ([]byte, error) { return xml.Marshal(v) }

func Unmarshal(r io.Reader, v any) error { return xml.NewDecoder(r).Decode(v) }
