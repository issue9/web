// SPDX-License-Identifier: MIT

// Package errs 对错误信息的二次处理
package errs

import (
	"github.com/issue9/localeutil"
	"golang.org/x/text/message"
)

type mergeErrors struct {
	err error
	msg localeutil.LocaleStringer
}

// Merge 合并多个错误实例
//
// 当两个值都不为 nil 时，会将 origin 作为其底层的错误类型，否则返回其中不为 nil 的值。
func Merge(origin, err error) error {
	if err == nil {
		return origin
	}

	if origin == nil {
		return err
	}

	return &mergeErrors{
		err: origin,
		msg: localeutil.Phrase("%s when return %s", err, origin),
	}
}

func (err *mergeErrors) Error() string { return err.LocaleString(localeutil.EmptyPrinter()) }

func (err *mergeErrors) Unwrap() error { return err.err }

func (err *mergeErrors) LocaleString(p *message.Printer) string { return err.msg.LocaleString(p) }
