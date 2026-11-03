// SPDX-FileCopyrightText: 2022-2026 caixw
//
// SPDX-License-Identifier: MIT

package validators

import "cmp"

// Between 创建判断数值区间 (min, max) 的验证器
func Between[T cmp.Ordered](min, max T) Validator[T] {
	if max < min {
		panic("max 必须大于等于 min")
	}
	return func(val T) bool { return val > min && val < max }
}

// BetweenEqual 创建判断数值区间 [min, max] 的验证器
func BetweenEqual[T cmp.Ordered](min, max T) Validator[T] {
	if max < min {
		panic("max 必须大于等于 min")
	}
	return func(val T) bool { return val >= min && val <= max }
}

// Less 创建判断值是否小于 num 的验证器
func Less[T cmp.Ordered](num T) Validator[T] { return func(t T) bool { return t < num } }

// Less 创建判断值是否小于或等于 num 的验证器
func LessEqual[T cmp.Ordered](num T) Validator[T] { return func(t T) bool { return t <= num } }

// Less 创建判断值是否大于 num 的验证器
func Great[T cmp.Ordered](num T) Validator[T] { return func(t T) bool { return t > num } }

// Less 创建判断值是否大于或等于 num 的验证器
func GreatEqual[T cmp.Ordered](num T) Validator[T] { return func(t T) bool { return t >= num } }

// ZeroOr 创建判断值为零值或是非零情况下符合 v 的要求的验证器
func ZeroOr[T comparable](v Validator[T]) Validator[T] {
	var zero T
	return Or(func(v T) bool { return v == zero }, v)
}
