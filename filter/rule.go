// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import "github.com/issue9/localeutil"

// Rule 过滤器的过滤规则
//
// name 为字段名；
// value 为需要验证的值指针；
// 返回字段名和错误信息，如果验证成功，则返回两个空值；
//
// Rule 合并了 [SanitizeFunc] 和 [Validator] 两者的功能。
type Rule[T any] func(name string, value *T) (string, localeutil.Stringer)

// Rule 将多个 [Rule] 合并为单个 [Rule]
func Rules[T any](rule ...Rule[T]) Rule[T] {
	return func(name string, value *T) (string, localeutil.Stringer) {
		for _, r := range rule {
			if name, msg := r(name, value); msg != nil {
				return name, msg
			}
		}
		return "", nil
	}
}

// ApplyRule 应用 rule 在 value 上
//
// 该方法是同时有 [Filter.Sanitize] 和 [Filter.Valid] 的功能。
func (f *Filter) ApplyRule[T any](name string, value *T, rule ...Rule[T]) *Filter {
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
