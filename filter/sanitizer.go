// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"github.com/issue9/localeutil"

	"github.com/issue9/web/filter/sanitizers"
)

// S2R 将一组 [sanitizers.Sanitizer] 转换为 [Rule]
func S2R[T any](s ...sanitizers.Sanitizer[T]) Rule[T] {
	if len(s) == 0 {
		panic("未指定参数")
	}

	var ss sanitizers.Sanitizer[T]
	if len(s) > 1 {
		ss = sanitizers.Sanitizers(s...)
	} else {
		ss = s[0]
	}

	return func(_ string, v *T) (string, localeutil.Stringer) {
		ss(v)
		return "", nil
	}
}

// SS2R 将一组用以修正切片元素的 [sanitizers.Sanitizer] 转换为 [Rule]
func SS2R[S ~[]T, T any](f ...sanitizers.Sanitizer[T]) Rule[S] {
	if len(f) == 0 {
		panic("未指定参数")
	}

	return func(_ string, s *S) (string, localeutil.Stringer) {
		for _, ff := range f {
			for index, item := range *s {
				ff(&item)
				(*s)[index] = item
			}
		}
		return "", nil
	}
}

// MS2R 将一组用以修正 map 元素的 [sanitizers.Sanitizer] 转换为 [Rule]
func MS2R[M ~map[K]V, K comparable, V any](f ...sanitizers.Sanitizer[V]) Rule[M] {
	if len(f) == 0 {
		panic("未指定参数")
	}

	return func(_ string, m *M) (string, localeutil.Stringer) {
		for _, ff := range f {
			for key, val := range *m {
				ff(&val)
				(*m)[key] = val
			}
		}
		return "", nil
	}
}

// Sanitize 通过 s 调整数据 value
//
// 当 [Filter.HasError] 和 [Filter.ExitOnError] 同时为真时，此方法也将不再调用。
func (f *Filter) Sanitize[T any](value *T, s ...sanitizers.Sanitizer[T]) *Filter {
	if value == nil {
		panic("参数 value 不能为空")
	}

	if f.HasError() && f.ExitOnError() {
		return f
	}

	for _, r := range s {
		r(value)
	}

	return f
}
