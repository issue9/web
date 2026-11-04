// SPDX-FileCopyrightText: 2022-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package validators 提供一些常用的验证器
package validators

import (
	"reflect"
	"regexp"
	"slices"
)

// Validator 验证器
type Validator[T any] = func(T) bool

// And 以与的形式串联多个验证器函数
func And[T any](v ...Validator[T]) Validator[T] {
	return func(val T) bool {
		for _, validator := range v {
			if !validator(val) {
				return false
			}
		}
		return true
	}
}

// Or 以或的形式并联多个验证器函数
func Or[T any](v ...Validator[T]) Validator[T] {
	return func(val T) bool {
		for _, validator := range v {
			if validator(val) {
				return true
			}
		}
		return false
	}
}

// Not 验证器的取反
func Not[T any](v Validator[T]) Validator[T] { return func(val T) bool { return !v(val) } }

// ZeroAny 是否为零值
//
// 如果是可比较类型，可以直接使用 [Zero] 或是使用 `Equal(0)` 创建一个验证器，
// 性能上会更好一些。
func ZeroAny[T any](v T) bool {
	rv := reflect.ValueOf(v)
	return !rv.IsValid() || rv.IsZero()
}

// Zero 是否为零值
//
// 相对于 [ZeroAny]，此方法性能稍好一些。但是对类型作了约束。
func Zero[T comparable](v T) bool {
	var zero T
	return zero == v
}

// Equal 生成判断值是否等于 v 的验证器
func Equal[T comparable](v T) Validator[T] { return func(t T) bool { return t == v } }

var nilKinds = []reflect.Kind{
	reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice,
}

// Nil 是否为 nil
func Nil[T any](v T) bool {
	rv := reflect.ValueOf(v)
	return !rv.IsValid() || (slices.Contains(nilKinds, rv.Kind()) && rv.IsNil())
}

// Regexp 创建一个正则表达式的验证器
//
// 如果 v 不是一个有效的正则表达式，则会 panic。
func Regexp(v string) Validator[string] {
	re, err := regexp.Compile(v)
	if err != nil {
		panic(err)
	}

	return func(s string) bool { return re.MatchString(s) }
}
