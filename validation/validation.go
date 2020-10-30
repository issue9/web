// SPDX-License-Identifier: MIT

// Package validation 数据验证相关功能
package validation

import "github.com/issue9/web/context"

// 当验证出错时的几种可用处理方式
const (
	ContinueAtError  ErrorHandling = iota // 碰到错误不中断验证
	ExitAtError                           // 碰到错误中断验证
	ExitFieldAtError                      // 碰到错误中断当前字段的验证
)

// ErrorHandling 当验证出错时的处理方式
type ErrorHandling int8

// Validation 验证器
type Validation struct {
	ctx           *context.Context
	errorHandling ErrorHandling
	errors        context.ResultFields
}

// New 返回新的 Validation 实例
//
// exitAtError 当验证器返回错误时，是否直接中断验证。
func New(ctx *context.Context, errorHandling ErrorHandling) *Validation {
	return &Validation{
		ctx:           ctx,
		errorHandling: errorHandling,
		errors:        context.ResultFields{},
	}
}

// NewField 验证新的字段
func (v *Validation) NewField(val interface{}, name string, rules ...Ruler) *Validation {
	if len(v.errors) > 0 && v.errorHandling == ExitAtError {
		return v
	}

	for _, rule := range rules {
		if msg := rule.Validate(val); msg != "" {
			v.errors.Add(name, msg)

			if v.errorHandling != ContinueAtError {
				return v
			}
		}
	}

	if len(v.errors[name]) > 0 { // 当前验证规则有错，则不验证子元素。
		return v
	}

	if vv, ok := val.(context.Validator); ok {
		if errors := vv.Validate(v.ctx); len(errors) > 0 {
			for key, vals := range errors {
				v.errors.Add(name+key, vals...)
			}
		}
	}

	return v
}

// Result 返回验证结果
func (v *Validation) Result() context.ResultFields {
	return v.errors
}