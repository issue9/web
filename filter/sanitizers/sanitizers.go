// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

// Package sanitizers 提供一些常用的修正器
package sanitizers

// Sanitizer 修正器
type Sanitizer[T any] = func(*T)

// Sanitizers 将多个 [Sanitizer] 合并为一个
func Sanitizers[T any](s ...Sanitizer[T]) Sanitizer[T] {
	return func(v *T) {
		for _, ss := range s {
			ss(v)
		}
	}
}

// None 不作任何修改
func None[T any](v *T) {}
