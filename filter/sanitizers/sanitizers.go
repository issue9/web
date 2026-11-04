// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

// Package sanitizers 提供一些常用的修正器
package sanitizers

import (
	"html"
	"strings"
	"unicode"
)

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

// Trim 过滤左右空格
func Trim(v *string) { *v = strings.TrimSpace(*v) }

func TrimLeft(v *string) {
	*v = strings.TrimLeftFunc(*v, func(r rune) bool { return unicode.IsSpace(r) })
}

func TrimRight(v *string) {
	*v = strings.TrimRightFunc(*v, func(r rune) bool { return unicode.IsSpace(r) })
}

func Upper(v *string) { *v = strings.ToUpper(*v) }

func Lower(v *string) { *v = strings.ToLower(*v) }

func EscapeHTML(v *string) { *v = html.EscapeString(*v) }

func UnescapeHTML(v *string) { *v = html.UnescapeString(*v) }
