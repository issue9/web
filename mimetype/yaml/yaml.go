// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

// Package yaml 支持 YAML 编码的序列化操作
//
// NOTE: 大部分时候，可以直接复用 json 标签，但是在对待嵌入对象时处理方式是不同的。
// json 自动展开，而 yaml 则需要指定 `yaml:",inline"` 才会展开。
package yaml

import (
	"io"

	"github.com/goccy/go-yaml"
	"github.com/issue9/mux/v9/header"

	"github.com/issue9/web"
)

const (
	Mimetype        = header.YAML
	ProblemMimetype = "application/problem+yaml"
)

func Marshal(_ *web.Context, v any) ([]byte, error) { return yaml.Marshal(v) }

func Unmarshal(r io.Reader, v any) error { return yaml.NewDecoder(r).Decode(v) }
