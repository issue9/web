// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"sync"

	"github.com/issue9/web/filter"
)

var filterContextPool = &sync.Pool{New: func() any { return &FilterContext{} }}

// Filter 用户提交数据的验证修正接口
type Filter interface {
	Filter(*FilterContext)
}

// FilterContext 处理由过滤器生成的各错误
type FilterContext struct {
	name        string // 当前检测对象的名称，如果是顶级对象应该是个空值。
	exitAtError bool
	ctx         *Context
	problem     *Problem
}

// NewFilterContext 声明 [FilterContext] 对象
func (ctx *Context) NewFilterContext(exitAtError bool) *FilterContext {
	return newFilterContext(exitAtError, "", ctx, newProblem())
}

// New 声明验证的子对象
//
// name 为 f 中验证对象的整体名称；
// f 为验证方法，其原型为 func(c *FilterContext)
// 往 c 参数写入的信息，其字段名均会以 name 作为前缀写入到当前对象 v 中。
// c 的各种属性均继承自 v。
func (v *FilterContext) New(name string, f func(c *FilterContext)) *FilterContext {
	f(newFilterContext(v.exitAtError, v.name+name, v.Context(), v.problem))
	return v
}

func newFilterContext(exitAtError bool, name string, ctx *Context, p *Problem) *FilterContext {
	v := filterContextPool.Get().(*FilterContext)
	v.name = name
	v.exitAtError = exitAtError
	v.ctx = ctx
	v.problem = p
	ctx.OnExit(func(*Context, int) { filterContextPool.Put(v) })
	return v
}

func (v *FilterContext) continueNext() bool { return !v.exitAtError || v.len() == 0 }

func (v *FilterContext) len() int { return len(v.problem.Params) }

// AddReason 直接添加一条错误信息
func (v *FilterContext) AddReason(name string, reason LocaleStringer) *FilterContext {
	if v.continueNext() {
		return v.addReason(name, reason)
	}
	return v
}

// AddError 直接添加一条类型为 [error] 的错误信息
func (v *FilterContext) AddError(name string, err error) *FilterContext {
	if ls, ok := err.(LocaleStringer); ok {
		return v.AddReason(name, ls)
	}
	return v.AddReason(name, Phrase(err.Error()))
}

func (v *FilterContext) addReason(name string, reason LocaleStringer) *FilterContext {
	if v.name != "" {
		name = v.name + name
	}
	v.problem.WithParam(name, reason.LocaleString(v.Context().LocalePrinter()))
	return v
}

// Add 添加由过滤器 f 返回的错误信息
func (v *FilterContext) Add(f filter.Filter) *FilterContext {
	if !v.continueNext() {
		return v
	}

	if name, msg := f(); msg != nil {
		v.addReason(name, msg)
	}
	return v
}

// AddFilter 验证实现了 [Filter] 接口的对象
func (v *FilterContext) AddFilter(name string, f Filter) *FilterContext {
	return v.New(name, func(fp *FilterContext) { f.Filter(fp) })
}

// When 只有满足 cond 才执行 f 中的验证
//
// f 中的 v 即为当前对象；
func (v *FilterContext) When(cond bool, f func(v *FilterContext)) *FilterContext {
	if cond {
		f(v)
	}
	return v
}

// Context 返回关联的 [Context] 实例
func (v *FilterContext) Context() *Context { return v.ctx }
