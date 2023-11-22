// SPDX-License-Identifier: MIT

// Package json JSON 格式的序列化方法
package json

import (
	"encoding/json"
	"io"

	"github.com/issue9/web"
)

const (
	Mimetype        = "application/json"
	ProblemMimetype = "application/problem+json"
)

func Marshal(_ *web.Context, v any) ([]byte, error) { return json.Marshal(v) }

func Unmarshal(r io.Reader, v any) error { return json.NewDecoder(r).Decode(v) }