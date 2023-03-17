// SPDX-License-Identifier: MIT

package filter

import (
	"fmt"
	"strconv"

	"github.com/issue9/localeutil"
)

type (
	// ValidatorFuncOf 验证器的函数原型
	ValidatorFuncOf[T any] func(T) bool

	// RulerFuncOf 数据验证规则
	//
	// 这是验证器与错误信息的组合。
	//
	// 传递参数为字段名与需要验证的值；
	// 返回字段名和错误信息，如果验证成功，则返回两个空值；
	RulerFuncOf[T any] func(string, T) (string, localeutil.LocaleStringer)
)

// Not 验证器的取反
func Not[T any](v func(T) bool) func(T) bool {
	return func(val T) bool { return !v(val) }
}

// AndFunc 以与的形式串联多个验证器函数
func And[T any](v ...func(T) bool) func(T) bool {
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
func Or[T any](v ...func(T) bool) func(T) bool {
	return func(val T) bool {
		for _, validator := range v {
			if validator(val) {
				return true
			}
		}
		return false
	}
}

func NewRule[T any](v func(T) bool, msg localeutil.LocaleStringer) RulerFuncOf[T] {
	return func(name string, val T) (string, localeutil.LocaleStringer) {
		if v(val) {
			return "", nil
		}
		return name, msg
	}
}

// NewRules 将多个规则合并为一个
//
// 按顺序依次验证，直接碰到第一个验证不过的。
func NewRules[T any](r ...RulerFuncOf[T]) RulerFuncOf[T] {
	return func(name string, val T) (string, localeutil.LocaleStringer) {
		for _, rule := range r {
			if n, msg := rule(name, val); msg != nil {
				return n, msg
			}
		}
		return "", nil
	}
}

// NewSliceRule 声明用于验证切片元素的规则
func NewSliceRule[T any, S ~[]T](v func(T) bool, msg localeutil.LocaleStringer) RulerFuncOf[S] {
	return func(name string, val S) (string, localeutil.LocaleStringer) {
		for index, vv := range val {
			if !v(vv) {
				return name + "[" + strconv.Itoa(index) + "]", msg
			}
		}
		return "", nil
	}
}

func NewSliceRules[T any, S ~[]T](r ...RulerFuncOf[T]) RulerFuncOf[S] {
	return func(name string, val S) (string, localeutil.LocaleStringer) {
		for _, rule := range r {
			for index, item := range val {
				if _, msg := rule(name, item); msg != nil {
					return name + "[" + strconv.Itoa(index) + "]", msg
				}
			}
		}
		return "", nil
	}
}

// NewMapRule 声明用于验证 map 元素的规则
func NewMapRule[K comparable, V any, M ~map[K]V](v func(V) bool, msg localeutil.LocaleStringer) RulerFuncOf[M] {
	return func(name string, val M) (string, localeutil.LocaleStringer) {
		for key, vv := range val {
			if !v(vv) {
				return fmt.Sprintf("%s[%v]", name, key), msg
			}
		}
		return "", nil
	}
}

func NewMapRules[K comparable, V any, M ~map[K]V](r ...RulerFuncOf[V]) RulerFuncOf[M] {
	return func(name string, val M) (string, localeutil.LocaleStringer) {
		for _, rule := range r {
			for key, item := range val {
				if _, msg := rule(name, item); msg != nil {
					return fmt.Sprintf("%s[%v]", name, key), msg
				}
			}
		}
		return "", nil
	}
}
