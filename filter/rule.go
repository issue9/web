// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"fmt"
	"strconv"

	"github.com/issue9/localeutil"
)

// Rule 过滤器使用的规则类型
//
// value 为需要验证的值指针；
// 返回字段名和错误信息，如果成功，则返回两个空值；
//
// 同时包含了 [sanitizers.Sanitizer] 和 [validators.Validator] 的功能。
// 大部分时候也都是通过 [V2R] 和 [S2R] 等方法将现有的
// [sanitizers.Sanitizer] 和 [validators.Validator] 组合成 [Rule] 供 [Filter.Apply] 使用。
type Rule[T any] func(value *T) localeutil.Stringer

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
		if ls := r(value); ls != nil {
			f.addReason(name, ls)
			break
		}
	}

	return f
}

// ApplySlice 对 value 的每一个元素使用 rule 规则
func (f *Filter) ApplySlice[S ~[]T, T any](name string, value *S, rule ...Rule[T]) *Filter {
	if value == nil {
		panic("参数 value 不能为空")
	}

	if f.HasError() && f.ExitOnError() {
		return f
	}

	for index, val := range *value {
		for _, r := range rule {
			if ls := r(&val); ls != nil {
				f.addReason(name+"["+strconv.Itoa(index)+"]", ls)

				if f.HasError() && f.ExitOnError() {
					return f
				}
			} else {
				// rule 只是修改了 val 的值，依然要将 val 重新写入 value
				(*value)[index] = val
			}
		}
	}

	return f
}

// ApplyMap 对 value 的每一个元素使用 rule 规则
func (f *Filter) ApplyMap[M ~map[K]V, K comparable, V any](name string, value M, rule ...Rule[V]) *Filter {
	if value == nil {
		panic("参数 value 不能为空")
	}

	if f.HasError() && f.ExitOnError() {
		return f
	}

	for key, val := range value {
		for _, r := range rule {
			if ls := r(&val); ls != nil {
				f.addReason(fmt.Sprintf("%s[%v]", name, key), ls)

				if f.HasError() && f.ExitOnError() {
					return f
				}
			} else {
				// rule 只是修改了 val 的值，依然要将 val 重新写入 value
				value[key] = val
			}
		}
	}

	return f
}
