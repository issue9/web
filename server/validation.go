// SPDX-License-Identifier: MIT

package server

import (
	"reflect"
	"strconv"
	"sync"

	"github.com/issue9/localeutil"
)

const validationPoolMaxSize = 20

var validationPool = &sync.Pool{New: func() any {
	const size = 5
	return &Validation{
		keys:    make([]string, 0, size),
		reasons: make([]string, 0, size),
	}
}}

var (
	isNotSlice = localeutil.Phrase("the type is not slice or array")
	isNotMap   = localeutil.Phrase("the type is not map")
)

type (
	// Validation 数据验证工具
	Validation struct {
		exitAtError bool
		ctx         *Context
		keys        []string
		reasons     []string
	}

	Rule struct {
		validator Validator
		message   localeutil.LocaleStringer
	}

	// Validator 用于验证指定数据的合法性
	Validator interface {
		IsValid(v any) bool
	}

	// ValidateFunc 用于验证指定数据的合法性
	ValidateFunc func(any) bool
)

func (f ValidateFunc) IsValid(v any) bool { return f(v) }

func (ctx *Context) NewValidation(exitAtError bool) *Validation {
	v := validationPool.Get().(*Validation)
	v.exitAtError = exitAtError
	v.keys = v.keys[:0]
	v.reasons = v.reasons[:0]
	v.ctx = ctx
	ctx.OnExit(func(i int) {
		if len(v.keys) < validationPoolMaxSize {
			validationPool.Put(v)
		}
	})
	return v
}

func (v *Validation) continueNext() bool { return !v.exitAtError || v.Count() == 0 }

func (v *Validation) Count() int { return len(v.keys) }

// Add 直接添加一条错误信息
func (v *Validation) Add(name string, reason localeutil.LocaleStringer) *Validation {
	if v.Count() > 0 && v.exitAtError {
		return v
	}
	return v.add(name, reason)
}

func (v *Validation) add(name string, reason localeutil.LocaleStringer) *Validation {
	v.keys = append(v.keys, name)
	v.reasons = append(v.reasons, reason.LocaleString(v.Context().LocalePrinter()))
	return v
}

// AddField 验证新的字段
//
// val 表示需要被验证的值；
// name 表示当前字段的名称，当验证出错时，以此值作为名称返回给用户；
// rules 表示验证的规则，按顺序依次验证。
func (v *Validation) AddField(val any, name string, rules ...*Rule) *Validation {
	if v.Count() > 0 && v.exitAtError {
		return v
	}

	for _, rule := range rules {
		if !rule.validator.IsValid(val) {
			v.add(name, rule.message)
			break
		}
	}
	return v
}

// AddSliceField 验证数组字段
//
// 如果字段类型不是数组或是字符串，将添加一条错误信息，并退出验证。
func (v *Validation) AddSliceField(val any, name string, rules ...*Rule) *Validation {
	// TODO: 如果 go 支持泛型方法，那么可以将 val 固定在 []T

	if v.Count() > 0 && v.exitAtError {
		return v
	}

	rv := reflect.ValueOf(val)

	if kind := rv.Kind(); kind != reflect.Array && kind != reflect.Slice && kind != reflect.String {
		v.add(name, isNotSlice)
		return v
	}

	for i := 0; i < rv.Len(); i++ {
		for _, rule := range rules {
			if !rule.validator.IsValid(rv.Index(i).Interface()) {
				v.add(name+"["+strconv.Itoa(i)+"]", rule.message)
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
// 如果字段类型不是 map，将添加一条错误信息，并退出验证。
func (v *Validation) AddMapField(val any, name string, rules ...*Rule) *Validation {
	// TODO: 如果 go 支持泛型方法，那么可以将 val 固定在 map[T]T

	if v.Count() > 0 && v.exitAtError {
		return v
	}

	rv := reflect.ValueOf(val)
	if kind := rv.Kind(); kind != reflect.Map {
		v.add(name, isNotMap)
		return v
	}

	keys := rv.MapKeys()
	for i := 0; i < rv.Len(); i++ {
		key := keys[i]
		for _, rule := range rules {
			if !rule.validator.IsValid(rv.MapIndex(key).Interface()) {
				v.add(name+"["+key.String()+"]", rule.message)
				if v.exitAtError {
					return v
				}
			}
		}
	}

	return v
}

func (v *Validation) Context() *Context { return v.ctx }

// When 只有满足 cond 才执行 f 中的验证
//
// f 中的 v 即为当前对象；
func (v *Validation) When(cond bool, f func(v *Validation)) *Validation {
	if cond {
		f(v)
	}
	return v
}

// AndValidator 将多个验证函数以与的形式合并为一个验证函数
func AndValidator(v ...Validator) Validator {
	return ValidateFunc(func(a any) bool {
		for _, validator := range v {
			if !validator.IsValid(a) {
				return false
			}
		}
		return true
	})
}

// OrValidator 将多个验证函数以或的形式合并为一个验证函数
func OrValidator(v ...Validator) Validator {
	return ValidateFunc(func(a any) bool {
		for _, validator := range v {
			if validator.IsValid(a) {
				return true
			}
		}
		return false
	})
}

func AndValidateFunc(f ...func(any) bool) Validator { return AndValidator(toValidators(f)...) }

func OrValidateFunc(f ...func(any) bool) Validator { return OrValidator(toValidators(f)...) }

func toValidators(f []func(any) bool) []Validator {
	v := make([]Validator, 0, len(f))
	for _, ff := range f {
		v = append(v, ValidateFunc(ff))
	}
	return v
}

// NewRule 声明一条验证规则
//
// message 表示在验证出错时的错误信息；
// validator 为验证方法；
func NewRule(message localeutil.LocaleStringer, validator Validator) *Rule {
	return &Rule{
		validator: validator,
		message:   message,
	}
}

func NewRuleFunc(message localeutil.LocaleStringer, f func(any) bool) *Rule {
	return NewRule(message, ValidateFunc(f))
}
