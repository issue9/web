// SPDX-License-Identifier: MIT

// Package serializer 序列化的相关操作
package serializer

import "github.com/issue9/localeutil"

var errUnsupported = localeutil.Error("unsupported serialization")

// ErrUnsupported 返回不支持序列化的错误信息
func ErrUnsupported() error { return errUnsupported }
