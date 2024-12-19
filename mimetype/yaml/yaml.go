// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package yaml 支持 YAML 编码的序列化操作
package yaml

import (
	"io"

	"gopkg.in/yaml.v3"

	"github.com/issue9/mux/v9/header"

	"github.com/issue9/web"
)

const (
	Mimetype        = header.YAML
	ProblemMimetype = "application/problem+yaml"
)

func Marshal(_ *web.Context, v any) ([]byte, error) { return yaml.Marshal(v) }

func Unmarshal(r io.Reader, v any) error { return yaml.NewDecoder(r).Decode(v) }
