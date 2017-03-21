// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"strings"

	"github.com/issue9/web/content"
)

// Renderer 向客户端渲染的接口
type Renderer interface {
	Render(status int, v interface{}, headers map[string]string)
}

// Reader 从客户端读取数据的接口
type Reader interface {
	Read(v interface{}) bool
}

// Context 是对 http.ResopnseWriter 和 http.Request 的简单封装。
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
	w http.ResponseWriter
	r *http.Request
	c content.Content
}

// NewContext 声明一个新的 Context
//
// 若 c 为空，则使用默认的内容。
func NewContext(w http.ResponseWriter, r *http.Request, c content.Content) *Context {
	if c == nil {
		c = defaultContent
	}

	return &Context{
		w: w,
		r: r,
		c: c,
	}
}

// Response 返回 http.ResoponseWriter 接口对象
func (ctx *Context) Response() http.ResponseWriter {
	return ctx.w
}

// Request 返回 *http.Request 对象
func (ctx *Context) Request() *http.Request {
	return ctx.r
}

// Render 将 v 渲染给客户端
func (ctx *Context) Render(status int, v interface{}, headers map[string]string) {
	ctx.c.Render(ctx.Response(), ctx.Request(), status, v, headers)
}

// Read 从客户端读取数据
func (ctx *Context) Read(v interface{}) bool {
	return ctx.c.Read(ctx.Response(), ctx.Request(), v)
}

// ParamID 获取地址参数中表示 ID 的值。相对于 int64，但该值必须大于 0。
// 当出错时，第二个参数返回 false。
func (ctx *Context) ParamID(key string, code int) (int64, bool) {
	p := ctx.Params()
	id := p.Int64(key)
	rslt := p.Result(code)

	if rslt.HasDetail() {
		rslt.Render(ctx)
		return id, false
	}

	if id <= 0 {
		rslt.Add("id", "必须大于零")
		rslt.Render(ctx)
		return id, false
	}

	return id, true
}

// ParamInt64 取地址参数中的 int64 值
func (ctx *Context) ParamInt64(key string, code int) (int64, bool) {
	p := ctx.Params()
	id := p.Int64(key)

	if p.OK(code) {
		return id, false
	}

	return id, true
}

// ResultFields 从报头中获取 X-Result-Fields 的相关内容。
//
// allow 表示所有允许出现的字段名称。
// 当第二个参数返回 true 时，返回的是可获取的字段名列表；
// 当第二个参数返回 false 时，返回的是不允许获取的字段名。
func (ctx *Context) ResultFields(allow []string) ([]string, bool) {
	resultFields := ctx.Request().Header.Get("X-Result-Fields")
	if len(resultFields) == 0 { // 没有指定，则返回所有字段内容
		return allow, true
	}
	fields := strings.Split(resultFields, ",")
	fails := make([]string, 0, len(fields))

	isAllow := func(field string) bool {
		for _, f1 := range allow {
			if f1 == field {
				return true
			}
		}
		return false
	}

	for index, field := range fields {
		field = strings.TrimSpace(field)
		fields[index] = field

		if !isAllow(field) { // 记录不允许获取的字段名
			fails = append(fails, field)
		}
	}

	if len(fails) > 0 {
		return fails, false
	}

	return fields, true
}
