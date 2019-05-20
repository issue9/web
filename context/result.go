// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import "github.com/issue9/web/app"

// Result 输出内容
type Result struct {
	rslt app.Result
	ctx  *Context
}

// NewResult 返回 Result 实例
func (ctx *Context) NewResult(code int) *Result {
	return &Result{
		rslt: ctx.App.NewResult(code),
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
func (rslt *Result) Add(key, val string) {
	rslt.rslt.Add(key, val)
}

// HasDetail 是否存在详细的错误信息
//
// 如果有通过 Add 添加内容，那么应该返回 true
func (rslt *Result) HasDetail() bool {
	return rslt.rslt.HasDetail()
}

// Render 渲染内容
func (rslt *Result) Render() {
	rslt.ctx.Render(rslt.rslt.Status(), rslt, nil)
}
