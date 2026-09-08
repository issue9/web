// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"fmt"
	"strconv"
)

// Rule 过滤器的过滤规则
//
// name 为字段名；
// value 为需要验证的值；
// 返回字段名和错误信息，如果验证成功，则返回两个空值；
type Rule[T any] func(name string, value *T) (string, LocaleStringer)

// SanitizerRule 将一组修正数据的函数封装为 [Rule]
func SanitizerRule[T any](f ...func(*T)) Rule[T] {
	return func(_ string, v *T) (string, LocaleStringer) {
		for _, ff := range f {
			ff(v)
		}
		return "", nil
	}
}

// ValidatorRule 将验证器函数封装为 [Rule]
func ValidatorRule[T any](f func(T) bool, msg LocaleStringer) Rule[T] {
	return func(name string, v *T) (string, LocaleStringer) {
		if !f(*v) {
			return name, msg
		}
		return "", nil
	}
}

// SliceSanitizerRule 将一组修正函数封装为 [Rule] 用以验证切片的元素
func SliceSanitizerRule[S ~[]T, T any](f ...func(*T)) Rule[S] {
	return func(_ string, s *S) (string, LocaleStringer) {
		for _, ff := range f {
			for index, item := range *s {
				ff(&item)
				(*s)[index] = item
			}
		}
		return "", nil
	}
}

// MapSanitizerRule 将一组修正函数封装为 [Rule] 用以验证 map 的元素
func MapSanitizerRule[M ~map[K]V, K comparable, V any](v func(*V)) Rule[M] {
	return func(_ string, m *M) (string, LocaleStringer) {
		for key, val := range *m {
			v(&val)
			(*m)[key] = val
		}
		return "", nil
	}
}

// SliceValidatorRule 将验证器封装为 [Rule] 用以验证切片元素
func SliceValidatorRule[S ~[]T, T any](v func(T) bool, msg LocaleStringer) Rule[S] {
	return func(name string, val *S) (string, LocaleStringer) {
		for index, vv := range *val {
			if !v(vv) {
				return name + "[" + strconv.Itoa(index) + "]", msg
			}
		}
		return "", nil
	}
}

// MapValidatorRule 将验证器封装为 [Rule] 用以验证 map
//
// v 用于验证键名和键值，两者可以有一个是空值，表示不需要验证，但不能都为空；
// msg 表示验证出错时的错误提示；
func MapValidatorRule[M ~map[K]V, K comparable, V any](v func(V) bool, msg LocaleStringer) Rule[M] {
	return func(name string, m *M) (string, LocaleStringer) {
		for key, val := range *m {
			if !v(val) {
				return fmt.Sprintf("%s[%v]", name, key), msg
			}
		}
		return "", nil
	}
}
