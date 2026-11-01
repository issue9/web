// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"fmt"
	"strconv"

	"github.com/issue9/localeutil"
)

// Rule 过滤器的过滤规则
//
// name 为字段名；
// value 为需要验证的值；
// 返回字段名和错误信息，如果验证成功，则返回两个空值；
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

// Sanitizer 将一组修正数据的函数封装为 [Rule]
func Sanitizer[T any](f ...func(*T)) Rule[T] {
	return func(_ string, v *T) (string, localeutil.Stringer) {
		for _, ff := range f {
			ff(v)
		}
		return "", nil
	}
}

// Validator 将验证器函数封装为 [Rule]
func Validator[T any](f func(T) bool, msg localeutil.Stringer) Rule[T] {
	return func(name string, v *T) (string, localeutil.Stringer) {
		if !f(*v) {
			return name, msg
		}
		return "", nil
	}
}

// SliceSanitizer 将一组修正函数封装为 [Rule] 用以验证切片的元素
func SliceSanitizer[S ~[]T, T any](f ...func(*T)) Rule[S] {
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

// MapSanitizer 将一组修正函数封装为 [Rule] 用以修正 map 的元素
func MapSanitizer[M ~map[K]V, K comparable, V any](f ...func(*V)) Rule[M] {
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

// SliceValidator 将验证器封装为 [Rule] 用以验证切片元素
func SliceValidator[S ~[]T, T any](v func(T) bool, msg localeutil.Stringer) Rule[S] {
	return func(name string, val *S) (string, localeutil.Stringer) {
		for index, vv := range *val {
			if !v(vv) {
				return name + "[" + strconv.Itoa(index) + "]", msg
			}
		}
		return "", nil
	}
}

// MapValidator 将验证器封装为 [Rule] 用以验证 map
//
// v 用于验证键值，两者可以有一个是空值，表示不需要验证，但不能都为空；
// msg 表示验证出错时的错误提示；
func MapValidator[M ~map[K]V, K comparable, V any](v func(V) bool, msg localeutil.Stringer) Rule[M] {
	return func(name string, m *M) (string, localeutil.Stringer) {
		for key, val := range *m {
			if !v(val) {
				return fmt.Sprintf("%s[%v]", name, key), msg
			}
		}
		return "", nil
	}
}
