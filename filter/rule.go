// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"fmt"
	"strconv"

	"github.com/issue9/localeutil"
)

// S 将一组修正数据的函数封装为 [Rule]
func S[T any](f ...func(*T)) Rule[T] {
	return func(_ string, v *T) (string, localeutil.Stringer) {
		for _, ff := range f {
			ff(v)
		}
		return "", nil
	}
}

// V 将验证器函数封装为 [Rule]
func V[T any](f func(T) bool, msg localeutil.Stringer) Rule[T] {
	return func(name string, v *T) (string, localeutil.Stringer) {
		if !f(*v) {
			return name, msg
		}
		return "", nil
	}
}

// SS 将一组修正函数封装为 [Rule] 用以验证切片的元素
func SS[S ~[]T, T any](f ...func(*T)) Rule[S] {
	return func(name string, s *S) (string, localeutil.Stringer) {
		for _, ff := range f {
			for index, item := range *s {
				ff(&item)
				(*s)[index] = item
			}
		}
		return "", nil
	}
}

// SS 将一组修正函数封装为 [Rule] 用以验证 map 的元素
func MS[M ~map[K]V, K comparable, V any](v func(*V)) Rule[M] {
	return func(name string, m *M) (string, localeutil.Stringer) {
		for key, val := range *m {
			v(&val)
			(*m)[key] = val
		}
		return "", nil
	}
}

// SV 将验证器封装为 [Rule] 用以验证切片元素
func SV[S ~[]T, T any](v func(T) bool, msg localeutil.Stringer) Rule[S] {
	return func(name string, val *S) (string, localeutil.Stringer) {
		for index, vv := range *val {
			if !v(vv) {
				return name + "[" + strconv.Itoa(index) + "]", msg
			}
		}
		return "", nil
	}
}

// MV 将验证器封装为 [Rule] 用以验证 map
//
// v 用于验证键名和键值，两者可以有一个是空值，表示不需要验证，但不能都为空；
// msg 表示验证出错时的错误提示；
func MV[M ~map[K]V, K comparable, V any](v func(V) bool, msg localeutil.Stringer) Rule[M] {
	return func(name string, m *M) (string, localeutil.Stringer) {
		for key, val := range *m {
			if !v(val) {
				return fmt.Sprintf("%s[%v]", name, key), msg
			}
		}
		return "", nil
	}
}
