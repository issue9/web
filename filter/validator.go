// SPDX-License-Identifier: MIT

package filter

// ValidatorFuncOf 验证器的函数原型
type ValidatorFuncOf[T any] func(T) bool

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
