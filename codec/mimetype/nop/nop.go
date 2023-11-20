// SPDX-License-Identifier: MIT

// Package nop 提供了空的序列化方法
package nop

import (
	"io"

	"github.com/issue9/web"
)

// Marshal 直接返回 [web.ErrUnsupportedSerialization]
func Marshal(*web.Context, any) ([]byte, error) { return nil, web.ErrUnsupportedSerialization() }

// Unmarshal 直接返回 [web.ErrUnsupportedSerialization]
func Unmarshal(io.Reader, any) error { return web.ErrUnsupportedSerialization() }
