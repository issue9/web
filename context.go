// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"

	"github.com/issue9/web/contentype"
	"github.com/issue9/web/request"
)

// Context 进程的上下文数据
//
//  ctx := web.NewContext(w, r)
//  id,ok := ctx.ParamID("id", 400001)
//  if !ok {
//      return
//  }
//
//  data := &Data{}
//  if !ctx.Read(data) {
//      // return
//  }
type Context struct {
	w  http.ResponseWriter
	r  *http.Request
	ct contentype.ContentTyper
}

// NewContext 声明一个新的 Context
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		w:  w,
		r:  r,
		ct: defaultContentType,
	}
}

// Render 将 v 渲染给客户端
func (ctx *Context) Render(status int, v interface{}, headers map[string]string) {
	ctx.ct.Render(ctx.w, ctx.r, status, v, headers)
}

// Read 从客户端读取数据
func (ctx *Context) Read(v interface{}) bool {
	return ctx.ct.Read(ctx.w, ctx.r, v)
}

// NewParam 声明一个新的 *request.Param 实例
func (ctx *Context) NewParam() *request.Param {
	return request.NewParam(ctx.r)
}

// NewQuery 声明一个新的 *request.Query 实例
func (ctx *Context) NewQuery() *request.Query {
	return request.NewQuery(ctx.r)
}

// ParamID 获取地址参数中表示 ID 的值。相对于 int64，但该值必须大于 0。
// 当出错时，第二个参数返回 false。
func (ctx *Context) ParamID(key string, code int) (int64, bool) {
	p := request.NewParam(ctx.r)
	id := p.Int64(key)
	rslt := p.Result(code)

	if rslt.HasDetail() {
		rslt.Render(ctx.w, ctx.r, ctx.ct)
		return id, false
	}

	if id <= 0 {
		rslt.Add("id", "必须大于零")
		rslt.Render(ctx.w, ctx.r, ctx.ct)
		return id, false
	}

	return id, true
}

// ParamInt64 取地址参数中的 int64 值
func (ctx *Context) ParamInt64(key string, code int) (int64, bool) {
	p := request.NewParam(ctx.r)
	id := p.Int64(key)

	if p.OK(ctx.w, ctx.r, code, ctx.ct) {
		return id, false
	}

	return id, true
}
