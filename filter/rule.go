// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import "github.com/issue9/localeutil"

// Rule 过滤器使用的规则类型
//
// name 为字段名；
// value 为需要验证的值指针；
// 返回字段名和错误信息，如果成功，则返回两个空值；
//
// 同时包含了 [sanitizers.Sanitizer] 和 [validators.Validator] 的功能。
// 大部分时候也都是通过 [V2R] 和 [S2R] 等方法将现有的
// [sanitizers.Sanitizer] 和 [validators.Validator] 组合成 [Rule] 供 [Filter.Apply] 使用。
type Rule[T any] func(name string, value *T) (string, localeutil.Stringer)

// Rules 将多个 [Rule] 合并为单个 [Rule]
func Rules[T any](rule ...Rule[T]) Rule[T] {
	return Rule[T](func(name string, value *T) (string, localeutil.Stringer) {
		for _, r := range rule {
			if name, msg := r(name, value); msg != nil {
				return name, msg
			}
		}
		return "", nil
	})
}

// Apply 对 value 使用 rule 规则
//
// 这是 [Filter] 的主要方法，用于对数据进行修正和验证。
// 如果只是单纯地修正数据或是验证数据，可以分别使用 [Filter.Sanitize] 和 [Filter.Valid]。
func (f *Filter) Apply[T any](name string, value *T, rule ...Rule[T]) *Filter {
	if value == nil {
		panic("参数 value 不能为空")
	}

	if f.HasError() && f.ExitOnError() {
		return f
	}

	for _, r := range rule {
		if name, ls := r(name, value); ls != nil {
			f.addReason(name, ls)
			break
		}
	}

	return f
}
