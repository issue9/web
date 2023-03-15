// SPDX-License-Identifier: MIT

package validation

import (
	"fmt"
	"strconv"

	"github.com/issue9/localeutil"
)

type (
	// RulerOf 数据验证规则
	//
	// 这是 [ValidatorOf] 与错误信息的组合。
	// 同时也负责将类型相关的泛型验证器转换成与类型无关的 [Field]。
	RulerOf[T any] interface {
		// Build 将参数与当前规则构建成 [Field] 对象
		Build(string, T) Field
	}

	RulerFuncOf[T any] func(string, T) Field
)

func (f RulerFuncOf[T]) Build(name string, v T) Field { return f(name, v) }

func NewRuleOf[T any](v ValidatorOf[T], msg localeutil.LocaleStringer) RulerOf[T] {
	return RulerFuncOf[T](func(name string, val T) Field {
		return FieldFunc(func() (string, localeutil.LocaleStringer) {
			if v.IsValid(val) {
				return "", nil
			}
			return name, msg
		})
	})
}

func NewRuleFuncOf[T any](v func(T) bool, msg localeutil.LocaleStringer) RulerOf[T] {
	return NewRuleOf[T](ValidatorFuncOf[T](v), msg)
}

// NewRulesOf 将多个规则合并为一个
//
// 按顺序依次验证，直接碰到第一个验证不过的。
func NewRulesOf[T any](r ...RulerOf[T]) RulerOf[T] {
	return RulerFuncOf[T](func(name string, val T) Field {
		return FieldFunc(func() (string, localeutil.LocaleStringer) {
			for _, rule := range r {
				if n, msg := rule.Build(name, val).Validate(); msg != nil {
					return n, msg
				}
			}
			return "", nil
		})
	})
}

// NewSliceRuleOf 声明用于验证切片元素的规则
func NewSliceRuleOf[T any, S ~[]T](v ValidatorOf[T], msg localeutil.LocaleStringer) RulerOf[S] {
	return RulerFuncOf[S](func(name string, val S) Field {
		return FieldFunc(func() (string, localeutil.LocaleStringer) {
			for index, vv := range val {
				if !v.IsValid(vv) {
					return name + "[" + strconv.Itoa(index) + "]", msg
				}
			}
			return "", nil
		})
	})
}

func NewSliceRulesOf[T any, S ~[]T](r ...RulerOf[T]) RulerOf[S] {
	return RulerFuncOf[S](func(name string, val S) Field {
		return FieldFunc(func() (string, localeutil.LocaleStringer) {
			for _, rule := range r {
				for index, item := range val {
					if _, msg := rule.Build(name, item).Validate(); msg != nil {
						return name + "[" + strconv.Itoa(index) + "]", msg
					}
				}
			}
			return "", nil
		})
	})
}

func NewSliceRuleFuncOf[T any, S ~[]T](v func(T) bool, msg localeutil.LocaleStringer) RulerOf[S] {
	return NewSliceRuleOf[T, S](ValidatorFuncOf[T](v), msg)
}

// NewMapRuleOf 声明用于验证 map 元素的规则
func NewMapRuleOf[K comparable, V any, M ~map[K]V](v ValidatorOf[V], msg localeutil.LocaleStringer) RulerOf[M] {
	return RulerFuncOf[M](func(name string, val M) Field {
		return FieldFunc(func() (string, localeutil.LocaleStringer) {
			for key, vv := range val {
				if !v.IsValid(vv) {
					return fmt.Sprintf("%s[%v]", name, key), msg
				}
			}
			return "", nil
		})
	})
}

func NewMapRuleFuncOf[K comparable, V any, M ~map[K]V](v func(V) bool, msg localeutil.LocaleStringer) RulerOf[M] {
	return NewMapRuleOf[K, V, M](ValidatorFuncOf[V](v), msg)
}

func NewMapRulesOf[K comparable, V any, M ~map[K]V](r ...RulerOf[V]) RulerOf[M] {
	return RulerFuncOf[M](func(name string, val M) Field {
		return FieldFunc(func() (string, localeutil.LocaleStringer) {
			for _, rule := range r {
				for key, item := range val {
					if _, msg := rule.Build(name, item).Validate(); msg != nil {
						return fmt.Sprintf("%s[%v]", name, key), msg
					}
				}
			}
			return "", nil
		})
	})
}
