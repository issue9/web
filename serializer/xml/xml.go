// SPDX-License-Identifier: MIT

// Package xml XML 编码的序列化操作
package xml

import (
	"encoding/xml"

	"github.com/issue9/web/server"
)

const (
	Mimetype        = "application/xml"
	ProblemMimetype = "application/problem+xml"
)

func Marshal(_ *server.Context, v any) ([]byte, error) { return xml.Marshal(v) }

func Unmarshal(data []byte, v any) error { return xml.Unmarshal(data, v) }
