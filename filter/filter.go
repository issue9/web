// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package filter 过滤器
//
// 包含了数据验证和数据修正两个功能。
//
// 各个类型之间的关系如下：
//
//	                      |---[Sanitize]
//	                      |
//	[Filter]---[FilterOf]-|---[RuleOf]-|---[LocaleStringer]
//	                                   |
//	                                   |---[Validator]
//
// Sanitize 表示对数据的修正，其函数原型为：func(*T)
// 指针传入数据，实现方可以对指向的数据进行修改。[sanitizer] 提供了一些通用的实现；
//
// Validator 负责验证数据，其原型为：func(T)bool
// 返回值表示是否符合当前函数的需求。[validator] 提供了一些通用的实现；
//
// [sanitizer]: https://pkg.go.dev/github.com/issue9/webfilter/sanitizer
// [validator]: https://pkg.go.dev/github.com/issue9/webfilter/validator
package filter

import "github.com/issue9/localeutil"

// Filter 执行过滤器的方法
//
// 当前方法由 [FilterOf] 生成，验证的数据也由其提供，
// 但是只有在调用当前方法时才真正对数据进行验证。
// 如果符合要求返回 "", nil，否则返回字段名和错误信息。
type Filter = func() (string, localeutil.Stringer)

// FilterOf 生成某数值的过滤器
//
// name 字段名，对于切片等类型会返回带下标的字段名；
// v 必须是指针类型，否则无法对其内容进行修改；
//
// 当前函数的主要作用是将一个泛型函数转换为非泛型函数 [Filter]。
type FilterOf[T any] func(name string, value *T) Filter

// RuleOf 对类型 T 的验证规则
//
// 传递参数为字段名与需要验证的值；
// 返回字段名和错误信息，如果验证成功，则返回两个空值；
type RuleOf[T any] func(string, *T) (string, localeutil.Stringer)

// New 声明过滤器
func New[T any](rule ...RuleOf[T]) FilterOf[T] {
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
