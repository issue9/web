// SPDX-License-Identifier: MIT

package problem

import (
	"reflect"
	"strconv"

	"github.com/issue9/localeutil"
	"golang.org/x/text/message"
)

type (
	Validation struct {
		exitAtError bool
		problems    *Problems
		params      []localeParam
	}

	localeParam struct {
		name   string
		reason localeutil.LocaleStringer
	}

	// Rule 验证规则
	//
	// 这是对 Validator 的二次包装，保存着未本地化的错误信息，用以在验证失败之后返回给 Validation。
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

// IsValid 将当前函数作为 Validator 使用
func (f ValidateFunc) IsValid(v any) bool { return f(v) }

func AndValidators(v ...Validator) Validator {
	return ValidateFunc(func(a any) bool {
		for _, validator := range v {
			if !validator.IsValid(a) {
				return false
			}
		}
		return true
	})
}

func OrValidators(v ...Validator) Validator {
	return ValidateFunc(func(a any) bool {
		for _, validator := range v {
			if validator.IsValid(a) {
				return true
			}
		}
		return false
	})
}

func FalseValidator(any) bool { return false }

func TrueValidator(any) bool { return true }

func NewRule(validator Validator, key message.Reference, v ...any) *Rule {
	return &Rule{
		validator: validator,
		message:   localeutil.Phrase(key, v...),
	}
}

// New 返回 Validation 对象
func (p *Problems) NewValidation() *Validation {
	return &Validation{
		exitAtError: p.exitAtError,
		problems:    p,
		params:      make([]localeParam, 0, 5),
	}
}

func (v *Validation) Problem(id string, p *message.Printer) Problem {
	if len(v.params) > 0 {
		pp := v.problems.Problem(id, p)
		for _, param := range v.params {
			pp.AddParam(param.name, param.reason.LocaleString(p))
		}
		return pp
	}
	return nil
}

func (v *Validation) Add(name string, reason localeutil.LocaleStringer) {
	v.params = append(v.params, localeParam{name: name, reason: reason})
}

// AddField 验证新的字段
//
// val 表示需要被验证的值；
// name 表示当前字段的名称，当验证出错时，以此值作为名称返回给用户；
// rules 表示验证的规则，按顺序依次验证。
func (v *Validation) AddField(val any, name string, rules ...*Rule) *Validation {
	if len(v.params) > 0 && v.problems.exitAtError {
		return v
	}

	for _, rule := range rules {
		if !rule.validator.IsValid(val) {
			v.Add(name, rule.message)
			break
		}
	}
	return v
}

// AddSliceField 验证数组字段
//
// 如果字段类型不是数组或是字符串，将直接返回错误。
func (v *Validation) AddSliceField(val any, name string, rules ...*Rule) *Validation {
	// TODO: 如果 go 支持泛型方法，那么可以将 val 固定在 []T

	rv := reflect.ValueOf(val)

	if kind := rv.Kind(); kind != reflect.Array && kind != reflect.Slice && kind != reflect.String {
		// 非数组，取第一个规则的错误信息。
		// TODO: 改成专门的错误信息
		v.Add(name, rules[0].message)
		return v
	}

	for i := 0; i < rv.Len(); i++ {
		for _, rule := range rules {
			if !rule.validator.IsValid(rv.Index(i).Interface()) {
				v.Add(name+"["+strconv.Itoa(i)+"]", rule.message)
				if v.exitAtError {
					return v
				}
			}
		}
	}

	return v
}

// AddMapField 验证 map 字段
//
// 如果字段类型不是 map，将直接返回错误。
func (v *Validation) AddMapField(val any, name string, rules ...*Rule) *Validation {
	// TODO: 如果 go 支持泛型方法，那么可以将 val 固定在 map[T]T

	rv := reflect.ValueOf(val)

	if kind := rv.Kind(); kind != reflect.Map {
		// 非数组，取第一个规则的错误信息。
		// TODO: 改成专门的错误信息
		v.Add(name, rules[0].message)
		return v
	}

	keys := rv.MapKeys()
	for i := 0; i < rv.Len(); i++ {
		key := keys[i]
		for _, rule := range rules {
			if !rule.validator.IsValid(rv.MapIndex(key).Interface()) {
				v.Add(name+"["+key.String()+"]", rule.message)
				if v.exitAtError {
					return v
				}
			}
		}
	}

	return v
}

// When 只有满足 cond 才执行 f 中的验证
//
// f 中的 v 即为当前对象；
func (v *Validation) When(cond bool, f func(v *Validation)) *Validation {
	if cond {
		f(v)
	}
	return v
}
