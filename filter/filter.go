// SPDX-License-Identifier: MIT

// Package filter 过滤器
//
// 包中的各个功能之间的关系如下：
//
//	                              |---[SanitizeFuncOf]
//	                              |
//	[FilterFunc]---[FilterFuncOf]-|---[RuleFuncOf]-|---[localeutil.LocaleStringer]
//	                                               |
//	                                               |---[ValidatorFuncOf]
//
// 调用者可以提前声明 [FilterFuncOf] 实例，在需要时调用 [FilterFuncOf] 实例，
// 生成一个类型无关的方法 [FilterFunc] 传递给 [server.FilterProblem]。
// 这样可以绕过 Go 不支持泛型方法的尴尬。
package filter

import "github.com/issue9/localeutil"

// FilterFunc 过滤器
//
// 当前方法由 [FilterFuncOf] 生成，验证的数据也有其提供，
// 但是只有在调用当前方法时才真正对数据进行验证。
// 如果符合要求返回 "", nil，否则返回字段名和错误信息。
type FilterFunc func() (string, localeutil.LocaleStringer)

// FilterFuncOf 生成某类型的过滤器
//
// name 字段名，对于切片等类型会返回带下标的字段名；
// v 必须是指针类型，否则无法对其内容进行修改；
type FilterFuncOf[T any] func(name string, v *T) FilterFunc

// NewFromVS 返回生成 [FilterFunc] 的方法
//
// msg 和 v 组成验证规则；
// s 表示对字段的一些清理，比如去除空白字符等，如果指定了此参数，会在 rule 之前执行；
func NewFromVS[T any](msg localeutil.LocaleStringer, v ValidatorFuncOf[T], s ...func(*T)) FilterFuncOf[T] {
	return New(NewRule(v, msg), s...)
}

// New 返回生成 [FilterFunc] 的方法
//
// sanitize 表示对字断 v 的一些清理，比如去除空白字符等，如果指定了此参数，会在 rule 之前执行；
func New[T any](rule RulerFuncOf[T], sanitize ...func(*T)) FilterFuncOf[T] {
	return func(name string, v *T) FilterFunc {
		return func() (string, localeutil.LocaleStringer) {
			for _, s := range sanitize {
				s(v)
			}

			if rule == nil {
				return "", nil
			}
			return rule(name, *v)
		}
	}
}
