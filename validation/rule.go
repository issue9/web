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
				return name, nil
			}
			return name, msg
		})
	})
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
			return name, nil
		})
	})
}

func NewSliceRuleOf[A []T, T any](v ValidatorOf[T], msg localeutil.LocaleStringer) RulerOf[A] {
	return RulerFuncOf[A](func(name string, val A) Field {
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

func NewMapRuleOf[A map[K]V, K comparable, V any](v ValidatorOf[V], msg localeutil.LocaleStringer) RulerOf[A] {
	return RulerFuncOf[A](func(name string, val A) Field {
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
