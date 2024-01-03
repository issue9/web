// SPDX-License-Identifier: MIT

// Package nop 提供了空的序列化方法
package nop

import (
	"io"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype"
)

// Marshal 直接返回 [mimetype.ErrUnsupported]
func Marshal(*web.Context, any) ([]byte, error) { return nil, mimetype.ErrUnsupported() }

// Unmarshal 直接返回 [mimetype.ErrUnsupported]
func Unmarshal(io.Reader, any) error { return mimetype.ErrUnsupported() }
