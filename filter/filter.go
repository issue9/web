// SPDX-License-Identifier: MIT

// Package filter 过滤器
//
// 提供了对数据的修正和验证功能。
package filter

import "github.com/issue9/localeutil"

// 如果符合要求返回 "", nil，否则返回错误的字段和信息。
type FilterFunc func() (string, localeutil.LocaleStringer)

// New 声明 Filter 对象
//
// sanitize 表示对字断 v 的一些清理，比如去除空白字符等，如果有多个可以使用 [Sanitizers] 进行合并；
func New[T any](name string, v *T, sanitize func(*T), rule RulerFuncOf[T]) FilterFunc {
	return func() (string, localeutil.LocaleStringer) {
		if sanitize != nil {
			sanitize(v)
		}
		return rule(name, *v)
	}
}
