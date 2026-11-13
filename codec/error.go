// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package codec

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

func (err errorUnsupported) Is(e error) bool { return e == errors.ErrUnsupported }

// ErrUnsupported 返回不支持序列化的错误信息
//
// 此方法也支持验证 [errors.ErrUnsupported]
//
//	errors.Is(ErrUnsupported(), errors.ErrUnsupported) == true。
func ErrUnsupported() error { return errUnsupported }
