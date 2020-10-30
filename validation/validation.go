// SPDX-License-Identifier: MIT

// Package validation 数据验证相关功能
package validation

import "github.com/issue9/query/v2"

// 当验证出错时的几种可用处理方式
const (
	ContinueAtError ErrorHandling = iota
	ExitAtError
	ExitFieldAtError
)

// ErrorHandling 当验证出错时的处理方式
type ErrorHandling int8

// Validation 验证器
type Validation struct {
	errorHandling ErrorHandling
	errors        query.Errors
}

// New 返回新的 Validation 实例
//
// exitAtError 当验证器返回错误时，是否直接中断验证。
func New(errorHandling ErrorHandling) *Validation {
	return &Validation{
		errorHandling: errorHandling,
		errors:        query.Errors{},
	}
}

// NewField 验证新的字段
func (v *Validation) NewField(val interface{}, name string, rules ...Ruler) *Validation {
	if len(v.errors) > 0 && v.errorHandling == ExitAtError {
		return v
	}

	for _, rule := range rules {
		if msg := rule.Validate(val); msg != "" {
			v.errors[name] = append(v.errors[name], msg)

			if v.errorHandling != ContinueAtError {
				return v
			}
		}
	}

	return v
}

// Result 返回结果
func (v *Validation) Result() query.Errors {
	return v.errors
}
