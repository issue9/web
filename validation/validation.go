// SPDX-License-Identifier: MIT

// Package validation 数据验证工具
package validation

import "github.com/issue9/localeutil"

// Field 待验证的字段需要实现的接口
type Field interface {
	// Validate 验证关联的数据
	//
	// 如果符合要求返回 "", nil，否则返回错误的字段和信息。
	Validate() (string, localeutil.LocaleStringer)
}

type FieldFunc func() (string, localeutil.LocaleStringer)

func (f FieldFunc) Validate() (string, localeutil.LocaleStringer) { return f() }
