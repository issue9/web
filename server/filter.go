// SPDX-License-Identifier: MIT

package server

import (
	"sync"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/filter"
)

const filterPoolMaxSize = 20

var filterPool = &sync.Pool{New: func() any {
	const size = 5
	return &Filter{
		keys:    make([]string, 0, size),
		reasons: make([]string, 0, size),
	}
}}

type Filter struct {
	exitAtError bool
	ctx         *Context
	keys        []string
	reasons     []string
}

// CTXFilter 在 [Context] 关联的上下文环境中对数据进行验证和修正
//
// 在 [Context.Read]、[Context.QueryObject] 以及 [Queries.Object]
// 中会在解析数据成功之后会调用该接口。
type CTXFilter interface {
	CTXFilter(*Filter)
}

func (ctx *Context) NewFilter(exitAtError bool) *Filter {
	v := filterPool.Get().(*Filter)
	v.exitAtError = exitAtError
	v.keys = v.keys[:0]
	v.reasons = v.reasons[:0]
	v.ctx = ctx
	ctx.OnExit(func(*Context, int) {
		if len(v.keys) < filterPoolMaxSize {
			filterPool.Put(v)
		}
	})
	return v
}

func (v *Filter) continueNext() bool { return !v.exitAtError || v.Len() == 0 }

func (v *Filter) Len() int { return len(v.keys) }

// Add 直接添加一条错误信息
func (v *Filter) Add(name string, reason localeutil.LocaleStringer) *Filter {
	if v.continueNext() {
		return v.add(name, reason)
	}
	return v
}

// AddError 直接添加一条类型为 error 的错误信息
func (v *Filter) AddError(name string, err error) *Filter {
	if ls, ok := err.(localeutil.LocaleStringer); ok {
		return v.Add(name, ls)
	}
	return v.Add(name, localeutil.Phrase(err.Error()))
}

func (v *Filter) add(name string, reason localeutil.LocaleStringer) *Filter {
	v.keys = append(v.keys, name)
	v.reasons = append(v.reasons, reason.LocaleString(v.Context().LocalePrinter()))
	return v
}

func (v *Filter) AddFilter(f filter.FilterFunc) *Filter {
	if !v.continueNext() {
		return v
	}

	if name, msg := f(); msg != nil {
		v.add(name, msg)
	}
	return v
}

// When 只有满足 cond 才执行 f 中的验证
//
// f 中的 v 即为当前对象；
func (v *Filter) When(cond bool, f func(v *Filter)) *Filter {
	if cond {
		f(v)
	}
	return v
}

// Context 返回关联的 [Context] 实例
func (v *Filter) Context() *Context { return v.ctx }

// Problem 转换成 [Problem] 对象
//
// 如果当前对象没有收集到错误，那么将返回 nil。
func (v *Filter) Problem(id string) Problem {
	if v == nil || v.Len() == 0 {
		return nil
	}

	p := v.Context().Problem(id)
	for index, key := range v.keys {
		p.AddParam(key, v.reasons[index])
	}
	return p
}
