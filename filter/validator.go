// SPDX-License-Identifier: MIT

package filter

import (
	"fmt"
	"strconv"

	"github.com/issue9/localeutil"
)

type (
	ValidatorFuncOf[T any] func(T) bool

	// RulerFuncOf 数据验证规则
	//
	// 这是 [ValidatorFuncOf] 与错误信息的组合。
	// 同时也负责将类型相关的泛型验证器转换成与类型无关的 [Field]。
	RulerFuncOf[T any] func(string, T) (string, localeutil.LocaleStringer)
)

// Not 这是对验证器的取反操作
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

func NewRuleOf[T any](v func(T) bool, msg localeutil.LocaleStringer) RulerFuncOf[T] {
	return func(name string, val T) (string, localeutil.LocaleStringer) {
		if v(val) {
			return "", nil
		}
		return name, msg
	}
}

// NewRulesOf 将多个规则合并为一个
//
// 按顺序依次验证，直接碰到第一个验证不过的。
func NewRulesOf[T any](r ...RulerFuncOf[T]) RulerFuncOf[T] {
	return func(name string, val T) (string, localeutil.LocaleStringer) {
		for _, rule := range r {
			if n, msg := rule(name, val); msg != nil {
				return n, msg
			}
		}
		return "", nil
	}
}

// NewSliceRuleOf 声明用于验证切片元素的规则
func NewSliceRuleOf[T any, S ~[]T](v func(T) bool, msg localeutil.LocaleStringer) RulerFuncOf[S] {
	return func(name string, val S) (string, localeutil.LocaleStringer) {
		for index, vv := range val {
			if !v(vv) {
				return name + "[" + strconv.Itoa(index) + "]", msg
			}
		}
		return "", nil
	}
}

func NewSliceRulesOf[T any, S ~[]T](r ...RulerFuncOf[T]) RulerFuncOf[S] {
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

// NewMapRuleOf 声明用于验证 map 元素的规则
func NewMapRuleOf[K comparable, V any, M ~map[K]V](v func(V) bool, msg localeutil.LocaleStringer) RulerFuncOf[M] {
	return func(name string, val M) (string, localeutil.LocaleStringer) {
		for key, vv := range val {
			if !v(vv) {
				return fmt.Sprintf("%s[%v]", name, key), msg
			}
		}
		return "", nil
	}
}

func NewMapRulesOf[K comparable, V any, M ~map[K]V](r ...RulerFuncOf[V]) RulerFuncOf[M] {
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
