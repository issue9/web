// SPDX-License-Identifier: MIT

package server

import (
	"sync"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/validation"
)

const validationPoolMaxSize = 20

var validationPool = &sync.Pool{New: func() any {
	const size = 5
	return &Validation{
		keys:    make([]string, 0, size),
		reasons: make([]string, 0, size),
	}
}}

type (
	// Validation 数据验证工具
	Validation struct {
		exitAtError bool
		ctx         *Context
		keys        []string
		reasons     []string
	}
)

// NewValidation 声明验证对象
func (ctx *Context) NewValidation(exitAtError bool) *Validation {
	v := validationPool.Get().(*Validation)
	v.exitAtError = exitAtError
	v.keys = v.keys[:0]
	v.reasons = v.reasons[:0]
	v.ctx = ctx
	ctx.OnExit(func(*Context, int) {
		if len(v.keys) < validationPoolMaxSize {
			validationPool.Put(v)
		}
	})
	return v
}

func (v *Validation) continueNext() bool { return !v.exitAtError || v.Len() == 0 }

func (v *Validation) Len() int { return len(v.keys) }

// Add 直接添加一条错误信息
func (v *Validation) Add(name string, reason localeutil.LocaleStringer) *Validation {
	if v.Len() > 0 && v.exitAtError {
		return v
	}
	return v.add(name, reason)
}

// AddError 直接添加一条类型为 err 的错误信息
func (v *Validation) AddError(name string, err error) *Validation {
	if ls, ok := err.(localeutil.LocaleStringer); ok {
		return v.Add(name, ls)
	}
	return v.Add(name, localeutil.Phrase(err.Error()))
}

func (v *Validation) add(name string, reason localeutil.LocaleStringer) *Validation {
	v.keys = append(v.keys, name)
	v.reasons = append(v.reasons, reason.LocaleString(v.Context().LocalePrinter()))
	return v
}

// AddField 验证新的字段
//
// name 表示当前字段的名称，当验证出错时，以此值作为名称返回给用户；
// validate 为验证方法，如果验证出错，则将返回错误信息，否则返回 nil；
func (v *Validation) AddField(validate validation.Field) *Validation {
	if v.Len() > 0 && v.exitAtError {
		return v
	}

	if name, msg := validate.Validate(); msg != nil {
		v.add(name, msg)
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

// Context 与当前验证对象关联的 [Context] 实例
func (v *Validation) Context() *Context { return v.ctx }
