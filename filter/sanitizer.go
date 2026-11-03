// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import "github.com/issue9/localeutil"

// Sanitizer 修正数据的函数签名
type Sanitizer[T any] = func(*T)

// S2R 将一组 [Sanitizer] 转换为 [Rule]
func S2R[T any](f ...Sanitizer[T]) Rule[T] {
	return func(_ string, v *T) (string, localeutil.Stringer) {
		for _, ff := range f {
			ff(v)
		}
		return "", nil
	}
}

// SS2R 将一组用以修正切片元素的 [Sanitizer] 转换为 [Rule]
func SS2R[S ~[]T, T any](f ...Sanitizer[T]) Rule[S] {
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

// SS2R 将一组用以修正 map 元素的 [Sanitizer] 转换为 [Rule]
func MS2R[M ~map[K]V, K comparable, V any](f ...Sanitizer[V]) Rule[M] {
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
func (f *Filter) Sanitize[T any](value *T, s ...Sanitizer[T]) *Filter {
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
