// SPDX-License-Identifier: MIT

// Package filter 过滤器
//
// 提供了对数据的修正和验证功能。
package filter

import "github.com/issue9/localeutil"

// FilterFunc 过滤器处理函数
//
// 如果符合要求返回 "", nil，否则返回错误的字段和信息。
type FilterFunc func() (string, localeutil.LocaleStringer)

// BuildFilterFuncOf 生成 [FilterFunc] 方法
//
// name 字段名，对于切片等类型会返回带下标的字段名；
// v 必须是指针类型，否则 sanitize 无法对其内容进行修改；
//
// 提供此方法，主要是为了将泛型转换非泛型方法。
type BuildFilterFuncOf[T any] func(name string, v *T) FilterFunc

// New 声明 FilterFunc 生成方法
//
// sanitize 表示对字断 v 的一些清理，比如去除空白字符等，如果指定了此参数，会在 rule 之前执行；
func New[T any](rule RulerFuncOf[T], sanitize ...func(*T)) BuildFilterFuncOf[T] {
	return func(name string, v *T) FilterFunc {
		return func() (string, localeutil.LocaleStringer) {
			for _, s := range sanitize {
				s(v)
			}

			if rule == nil {
				return "", nil
			}
			return rule(name, *v)
		}
	}
}
