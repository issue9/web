// SPDX-FileCopyrightText: 2018-2026 caixw
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
		logs.AppendError(p, b, t)
		return string(b.Bytes())
	case localeutil.Stringer:
		return t.LocaleString(p)
	default:
		return err.Error()
	}
}
