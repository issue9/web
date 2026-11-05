// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"fmt"
	"strconv"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/filter/validators"
)

// V2R 将 [validators.Validator] 转换为 [Rule]
func V2R[T any](msg localeutil.Stringer, f validators.Validator[T]) Rule[T] {
	return func(v *T) localeutil.Stringer {
		if !f(*v) {
			return msg
		}
		return nil
	}
}

// 通过函数 v 验证数据 value
//
// 如果无法通过验证，那么将以 name 作为字段名 msg 作为错误信息保存。
func (f *Filter) Valid[T any](name string, value T, v validators.Validator[T], msg localeutil.Stringer) *Filter {
	if f.HasError() && f.ExitOnError() {
		return f
	}

	if !v(value) {
		f.addReason(name, msg)
	}

	return f
}

// ValidSlice 验证 value 中的每一个元素
func (f *Filter) ValidSlice[S ~[]T, T any](name string, value S, v validators.Validator[T], msg localeutil.Stringer) *Filter {
	if f.HasError() && f.ExitOnError() {
		return f
	}

	for index, val := range value {
		if !v(val) {
			f.addReason(name+"["+strconv.Itoa(index)+"]", msg)

			if f.HasError() && f.ExitOnError() {
				return f
			}
		}
	}

	return f
}

// ValidMap 验证 value 中的每一个元素
func (f *Filter) ValidMap[M ~map[K]V, K comparable, V any](name string, value M, v validators.Validator[V], msg localeutil.Stringer) *Filter {
	if f.HasError() && f.ExitOnError() {
		return f
	}

	for key, val := range value {
		if !v(val) {
			f.addReason(fmt.Sprintf("%s[%v]", name, key), msg)

			if f.HasError() && f.ExitOnError() {
				return f
			}
		}
	}

	return f
}
