// SPDX-License-Identifier: MIT

package validation

import "github.com/issue9/localeutil"

type (
	Rule struct {
		validator Validator
		message   localeutil.LocaleStringer
	}

	// Validator 用于验证指定数据的合法性
	Validator interface {
		// IsValid 验证 v 是否符合当前的规则
		IsValid(v any) bool
	}

	// ValidateFunc 用于验证指定数据的合法性
	ValidateFunc func(any) bool
)

func (f ValidateFunc) IsValid(v any) bool { return f(v) }

// And 将多个验证函数以与的形式合并为一个验证函数
func And(v ...Validator) Validator {
	return ValidateFunc(func(a any) bool {
		for _, validator := range v {
			if !validator.IsValid(a) {
				return false
			}
		}
		return true
	})
}

// Or 将多个验证函数以或的形式合并为一个验证函数
func Or(v ...Validator) Validator {
	return ValidateFunc(func(a any) bool {
		for _, validator := range v {
			if validator.IsValid(a) {
				return true
			}
		}
		return false
	})
}

func AndF(f ...func(any) bool) Validator { return And(toValidators(f)...) }

func OrF(f ...func(any) bool) Validator { return Or(toValidators(f)...) }

func toValidators(f []func(any) bool) []Validator {
	v := make([]Validator, 0, len(f))
	for _, ff := range f {
		v = append(v, ValidateFunc(ff))
	}
	return v
}

func NewRule(message localeutil.LocaleStringer, validator Validator) *Rule {
	return &Rule{
		validator: validator,
		message:   message,
	}
}

func NewRuleFunc(message localeutil.LocaleStringer, f func(any) bool) *Rule {
	return NewRule(message, ValidateFunc(f))
}
