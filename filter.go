// SPDX-License-Identifier: MIT

package web

import (
	"fmt"
	"strconv"
)

// FilterFunc 过滤器
//
// 当前方法由 [FilterFuncOf] 生成，验证的数据也由其提供，
// 但是只有在调用当前方法时才真正对数据进行验证。
// 如果符合要求返回 "", nil，否则返回字段名和错误信息。
//
// FilterFunc 与其它各个的关系如下：
//
//	                              |---[Sanitize]
//	                              |
//	[FilterFunc]---[FilterFuncOf]-|---[RuleFuncOf]-|---[localeutil.LocaleStringer]
//	                                               |
//	                                               |---[Validator]
//
// 调用者可以提前声明 [FilterFuncOf] 实例，在需要时调用 [FilterFuncOf] 实例，
// 生成一个类型无关的方法 [FilterFunc] 传递给 [web.FilterProblem]。
// 这样可以绕过 Go 不支持泛型方法的尴尬。
//
// # Sanitize
//
// 数据修正发生成数据验证之前，其函数原型为：
//
//	func(*T)
//
// 指针传入数据，实现方可以对指向的数据进行修改。
// [sanitizer] 提供了一些通用的实现；
//
// # Validator
//
// 验证器只负责验证数据，其原型为：
//
//	func(T)bool
//
// 返回值表示是否符合当前函数的需求。
// [validator] 提供了一些通用的实现；
//
// # Rule
//
// 这是验证器与提示信息的结合，当不符合当前规则所包含的验证器需求时，返回对应的错误信息。
//
// [sanitizer]: https://pkg.go.dev/github.com/issue9/filter/sanitizer
// [validator]: https://pkg.go.dev/github.com/issue9/filter/validator
type FilterFunc func() (string, LocaleStringer)

// FilterFuncOf 生成某数值的过滤器
//
// name 字段名，对于切片等类型会返回带下标的字段名；
// v 必须是指针类型，否则无法对其内容进行修改；
type FilterFuncOf[T any] func(name string, v *T) FilterFunc

// RuleFuncOf 数据验证规则
//
// 这是验证器与错误信息的组合。
//
// 传递参数为字段名与需要验证的值；
// 返回字段名和错误信息，如果验证成功，则返回两个空值；
type RuleFuncOf[T any] func(string, T) (string, LocaleStringer)

// NewFilterFromVS 生成 [FilterFuncOf]
//
// msg 和 v 组成验证规则；
// s 表示对字段 v 的一些清理，比如去除空白字符等，如果指定了此参数，会在 rule 之前执行；
func NewFilterFromVS[T any](msg LocaleStringer, v func(T) bool, s ...func(*T)) FilterFuncOf[T] {
	return NewFilter(NewRule(v, msg), s...)
}

// NewFilter 生成 [FilterFuncOf]
//
// s 表示对字段 v 的一些清理，比如去除空白字符等，如果指定了此参数，会在 rule 之前执行；
func NewFilter[T any](rule RuleFuncOf[T], s ...func(*T)) FilterFuncOf[T] {
	return func(name string, v *T) FilterFunc {
		return func() (string, LocaleStringer) {
			for _, sa := range s {
				sa(v)
			}

			if rule == nil {
				return "", nil
			}
			return rule(name, *v)
		}
	}
}

// NewSliceFilter 生成针对切片元素的 [FilterFuncOf]
//
// rule 和 s 将会应用到每个元素 T 上。
func NewSliceFilter[T any, S ~[]T](rule RuleFuncOf[T], s ...func(*T)) FilterFuncOf[S] {
	r := NewSliceRules[T, []T](rule)

	return func(name string, v *S) FilterFunc {
		return func() (string, LocaleStringer) {
			for _, sa := range s {
				for index, item := range *v {
					sa(&item)
					(*v)[index] = item
				}
			}

			if rule == nil {
				return "", nil
			}
			return r(name, *v)
		}
	}
}

// NewMapFilter 生成针对 map 元素的 [FilterFuncOf]
//
// rule 和 s 将会应用到每个元素 T 上。
func NewMapFilter[K comparable, V any, M ~map[K]V](rule RuleFuncOf[V], s ...func(*V)) FilterFuncOf[M] {
	r := NewMapRules[K, V, map[K]V](rule)
	return func(name string, v *M) FilterFunc {
		return func() (string, LocaleStringer) {
			for _, sa := range s {
				for k, item := range *v {
					sa(&item)
					(*v)[k] = item
				}
			}

			if rule == nil {
				return "", nil
			}
			return r(name, *v)
		}
	}
}

func NewRule[T any](v func(T) bool, msg LocaleStringer) RuleFuncOf[T] {
	return func(name string, val T) (string, LocaleStringer) {
		if v(val) {
			return "", nil
		}
		return name, msg
	}
}

// NewRules 将多个规则合并为一个
//
// 按顺序依次验证，直接碰到第一个验证不过的。
func NewRules[T any](r ...RuleFuncOf[T]) RuleFuncOf[T] {
	return func(name string, val T) (string, LocaleStringer) {
		for _, rule := range r {
			if n, msg := rule(name, val); msg != nil {
				return n, msg
			}
		}
		return "", nil
	}
}

// NewSliceRule 声明用于验证切片元素的规则
func NewSliceRule[T any, S ~[]T](v func(T) bool, msg LocaleStringer) RuleFuncOf[S] {
	return func(name string, val S) (string, LocaleStringer) {
		for index, vv := range val {
			if !v(vv) {
				return name + "[" + strconv.Itoa(index) + "]", msg
			}
		}
		return "", nil
	}
}

func NewSliceRules[T any, S ~[]T](r ...RuleFuncOf[T]) RuleFuncOf[S] {
	return func(name string, val S) (string, LocaleStringer) {
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

// NewMapRule 声明验证 map 元素的规则
func NewMapRule[K comparable, V any, M ~map[K]V](v func(V) bool, msg LocaleStringer) RuleFuncOf[M] {
	return func(name string, val M) (string, LocaleStringer) {
		for key, vv := range val {
			if !v(vv) {
				return fmt.Sprintf("%s[%v]", name, key), msg
			}
		}
		return "", nil
	}
}

func NewMapRules[K comparable, V any, M ~map[K]V](r ...RuleFuncOf[V]) RuleFuncOf[M] {
	return func(name string, val M) (string, LocaleStringer) {
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
