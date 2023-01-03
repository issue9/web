// SPDX-License-Identifier: MIT

// Package json JSON 格式的序列化方法
package json

import (
	"encoding/json"

	"github.com/issue9/web/server"
)

const (
	Mimetype        = "application/json"
	ProblemMimetype = "application/problem+json"
)

func Marshal(_ *server.Context, v any) ([]byte, error) { return json.Marshal(v) }

func Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
