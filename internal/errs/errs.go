// SPDX-License-Identifier: MIT

// Package errs 与错误相关的定义
package errs

import "github.com/issue9/localeutil"

func NewLocaleError(format string, v ...any) error {
	return localeutil.Error(format, v...)
}
