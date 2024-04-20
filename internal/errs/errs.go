// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package errs 与错误相关的定义
package errs

import (
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v7"
	"golang.org/x/xerrors"
)

func Sprint(p *localeutil.Printer, err error, detail bool) string {
	if err == nil || p == nil {
		panic("参数 p 和 err 不能为 nil")
	}

	switch t := err.(type) {
	case xerrors.Formatter:
		b := logs.NewBuffer(detail)
		e := t.FormatError(b)
	FOR:
		for e != nil {
			switch tt := e.(type) {
			case xerrors.Formatter:
				e = tt.FormatError(b)
			case localeutil.Stringer:
				b.AppendString(tt.LocaleString(p))
				break FOR
			default:
				b.AppendString(tt.Error())
				break FOR
			}
		}
		return string(b.Bytes())
	case localeutil.Stringer:
		return t.LocaleString(p)
	default:
		return err.Error()
	}
}
