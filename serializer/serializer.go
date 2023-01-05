// SPDX-License-Identifier: MIT

// Package serializer 序列化的相关操作
package serializer

import (
	"errors"

	"github.com/issue9/web/errs"
)

var errUnsupported = errs.NewLocaleError("unsupported serialization")

// ErrUnsupported 返回不支持序列化的错误信息
func ErrUnsupported() error { return errUnsupported }

// IsUnsupported 判断 err 是否为 [ErrUnsupported] 返回的值
func IsUnsupported(err error) bool { return errors.Is(err, ErrUnsupported()) }
