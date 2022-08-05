// SPDX-License-Identifier: MIT

package validation

import (
	"github.com/issue9/localeutil"
	"golang.org/x/text/message"
)

type (
	// Rule 验证规则
	//
	// 与 [Validator] 相比，包含了本地化的错误信息。
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

func NewRule(validator Validator, key message.Reference, v ...any) *Rule {
	return &Rule{
		validator: validator,
		message:   localeutil.Phrase(key, v...),
	}
}
