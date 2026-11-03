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
	return func(name string, v *T) (string, localeutil.Stringer) {
		if !f(*v) {
			return name, msg
		}
		return "", nil
	}
}

// SV2R 将一组用以验证切片元素的 [validators.Validator] 转换为 [Rule]
func SV2R[S ~[]T, T any](msg localeutil.Stringer, f validators.Validator[T]) Rule[S] {
	return func(name string, val *S) (string, localeutil.Stringer) {
		for index, vv := range *val {
			if !f(vv) {
				return name + "[" + strconv.Itoa(index) + "]", msg
			}
		}
		return "", nil
	}
}

// MV2R 将一组用以验证 map 元素的 [validators.Validator] 转换为 [Rule]
func MV2R[M ~map[K]V, K comparable, V any](msg localeutil.Stringer, f validators.Validator[V]) Rule[M] {
	return func(name string, m *M) (string, localeutil.Stringer) {
		for key, val := range *m {
			if !f(val) {
				return fmt.Sprintf("%s[%v]", name, key), msg
			}
		}
		return "", nil
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
