// SPDX-License-Identifier: MIT

// Package mimetype 对媒体类型的编解码处理
package mimetype

import (
	"errors"

	"github.com/issue9/localeutil"
)

var errUnsupported = errorUnsupported{}

type errorUnsupported struct{}

func (err errorUnsupported) Error() string { return err.LocaleString(nil) }

func (err errorUnsupported) LocaleString(p *localeutil.Printer) string {
	return localeutil.StringPhrase("unsupported serialization").LocaleString(p)
}

func (err errorUnsupported) Unwrap() error { return errors.ErrUnsupported }

// ErrUnsupported 返回不支持序列化的错误信息
//
// 此方法的返回对象同时也包含了 [errors.ErrUnsupported]，
// errors.Is(ErrUnsupported(), errors.ErrUnsupported) == true。
func ErrUnsupported() error { return errUnsupported }
