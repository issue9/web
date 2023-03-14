// SPDX-License-Identifier: MIT

package validation

// ValidatorOf 验证器
type ValidatorOf[T any] interface {
	// IsValid 验证参数是否符合验证器的要求
	IsValid(T) bool
}

// ValidatorFuncOf 验证器的函数形式
type ValidatorFuncOf[T any] func(T) bool

func (f ValidatorFuncOf[T]) IsValid(v T) bool { return f(v) }

// Not 这是对 [ValidatorOf] 的取反操作
func Not[T any](v ValidatorOf[T]) ValidatorOf[T] {
	return ValidatorFuncOf[T](func(val T) bool { return !v.IsValid(val) })
}

// And 以与的形式串联多个验证器
func And[T any](v ...ValidatorOf[T]) ValidatorOf[T] {
	return ValidatorFuncOf[T](func(val T) bool {
		for _, validator := range v {
			if !validator.IsValid(val) {
				return false
			}
		}
		return true
	})
}

// AndFunc 以与的形式串联多个验证器函数
func AndFunc[T any](v ...func(T) bool) ValidatorOf[T] {
	return ValidatorFuncOf[T](func(val T) bool {
		for _, validator := range v {
			if !validator(val) {
				return false
			}
		}
		return true
	})
}

// Or 以或的形式并联多个验证器
func Or[T any](v ...ValidatorOf[T]) ValidatorOf[T] {
	return ValidatorFuncOf[T](func(val T) bool {
		for _, validator := range v {
			if validator.IsValid(val) {
				return true
			}
		}
		return false
	})
}

// OrFunc 以或的形式并联多个验证器函数
func OrFunc[T any](v ...func(T) bool) ValidatorOf[T] {
	return ValidatorFuncOf[T](func(val T) bool {
		for _, validator := range v {
			if validator(val) {
				return true
			}
		}
		return false
	})
}
