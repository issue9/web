// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package json JSON 格式的序列化方法
package json

import (
	"encoding/json/v2"
	"io"

	"github.com/issue9/mux/v9/header"

	"github.com/issue9/web"
)

const (
	Mimetype        = header.JSON
	ProblemMimetype = "application/problem+json"
)

func Marshal(_ *web.Context, v any) ([]byte, error) { return json.Marshal(v) }

func Unmarshal(r io.Reader, v any) error { return json.UnmarshalRead(r, v) }
