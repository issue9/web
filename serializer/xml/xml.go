// SPDX-License-Identifier: MIT

// Package xml XML 编码的序列化操作
package xml

import (
	"encoding/xml"

	"github.com/issue9/web"
)

const (
	Mimetype        = "application/xml"
	ProblemMimetype = "application/problem+xml"
)

func BuildMarshal(*web.Context) web.MarshalFunc { return xml.Marshal }

func Unmarshal(data []byte, v any) error { return xml.Unmarshal(data, v) }
