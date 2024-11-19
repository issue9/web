// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package filter 过滤器
//
// 包含了数据验证和数据修正两个功能。
//
// 各个类型之间的关系如下：
//
//	                           |---[Sanitize]
//	                           |
//	[Filter]---[Builder]----[Rule]
//	                           |
//	                           |---[Validator]
//
// Sanitize 表示对数据的修正，其函数原型为：func(*T)
// 指针传入数据，实现方可以对指向的数据进行修改，可由 [S]、[SS] 或 [MS] 转换为 [Rule]；
//
// Validator 负责验证数据，其原型为：func(T)bool
// 返回值表示是否符合当前函数的需求，可由 [V]、[SV] 或 [MV] 转换为 [Rule]；
package filter

import (
	"github.com/issue9/config"
	"github.com/issue9/localeutil"
)

// Filter 过滤器函数类型
//
// 当前方法由 [Builder] 生成，验证的数据也由其提供，
// 但是只有在调用当前方法时才真正对数据进行验证。
// 如果符合要求返回 "", nil，否则返回字段名和错误信息。
type Filter = func() (string, localeutil.Stringer)

// Builder 生成类型 T 的过滤器
//
// name 字段名，对于切片等类型会返回带下标的字段名；
// v 必须是指针类型，否则无法对其内容进行修改；
//
// 当前函数的主要作用是将一个泛型函数转换为非泛型函数 [Filter]。
type Builder[T any] func(name string, value *T) Filter

// Rule 类型 T 的验证规则
//
// 传递参数为字段名与需要验证的值；
// 返回字段名和错误信息，如果验证成功，则返回两个空值；
type Rule[T any] func(string, *T) (string, localeutil.Stringer)

// NewBuilder 声明 [Builder]
//
// 按参数的添加顺序依次执行。
func NewBuilder[T any](rule ...Rule[T]) Builder[T] {
	return func(name string, value *T) Filter {
		return func() (string, localeutil.Stringer) {
			for _, r := range rule {
				if name, ls := r(name, value); ls != nil {
					return name, ls
				}
			}
			return "", nil
		}
	}
}

// New 声明 [Filter]
//
// name 和 value 为调用 [Builder] 的参数；
// rule 为声明 [Builder] 的参数；
func New[T any](name string, value *T, rule ...Rule[T]) Filter {
	return NewBuilder(rule...)(name, value)
}

// ToFieldError 将 [Filter] 返回的错误转换为 [config.FieldError]
//
// 若所有过滤器都没有返回错误信息，则此方法返回 nil。
func ToFieldError(f ...Filter) *config.FieldError {
	for _, ff := range f {
		if name, msg := ff(); msg != nil {
			return config.NewFieldError(name, msg)
		}
	}
	return nil
}
