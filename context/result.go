// SPDX-License-Identifier: MIT

package context

import (
	"github.com/issue9/web/result"
)

// Result 输出内容
type Result struct {
	rslt result.Result
	ctx  *Context
}

// NewResult 返回 Result 实例
func (ctx *Context) NewResult(code int) *Result {
	return &Result{
		rslt: ctx.builder.Results().NewResult(code),
		ctx:  ctx,
	}
}

// NewResultWithDetail 返回 Result 实例
func (ctx *Context) NewResultWithDetail(code int, detail map[string]string) *Result {
	rslt := ctx.NewResult(code)

	for k, v := range detail {
		rslt.Add(k, v)
	}

	return rslt
}

// Add 添加详细的内容
func (rslt *Result) Add(key, val string) *Result {
	rslt.rslt.Add(key, val)
	return rslt
}

// HasDetail 是否存在详细的错误信息
//
// 如果有通过 Add 添加内容，那么应该返回 true
func (rslt *Result) HasDetail() bool {
	return rslt.rslt.HasDetail()
}

// Render 渲染内容
func (rslt *Result) Render() {
	rslt.ctx.Render(rslt.rslt.Status(), rslt.rslt, nil)
}
