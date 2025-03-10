// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

// Package toml [TOML] 编码
//
// [TOML]: https://toml.io/cn/
package toml

import (
	"io"

	"github.com/BurntSushi/toml"

	"github.com/issue9/web"
)

const (
	Mimetype        = "application/toml"
	ProblemMimetype = "application/problem+toml"
)

func Marshal(_ *web.Context, v any) ([]byte, error) { return toml.Marshal(v) }

func Unmarshal(r io.Reader, v any) error {
	_, err := toml.NewDecoder(r).Decode(v)
	return err
}
