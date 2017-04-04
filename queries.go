// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import "strconv"

// Queries 用于处理路径中的查询参数。用法类似于 flag
//  q,_ := NewQuery(r, false)
//  page := q.Int64("page", 1)
//  size := q.Int64("size", 20)
//  if r := q.Result();r.HasDetail(){
//      rslt.RenderJSON(w)
//      return
//  }
type Queries struct {
	ctx    *Context
	errors map[string]string
}

// Queries 声明一个新的 Queries 实例
func (ctx *Context) Queries() *Queries {
	return &Queries{
		ctx:    ctx,
		errors: map[string]string{},
	}
}

// Int 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
//
// 若是无法转换，则会保存错误信息
func (q *Queries) Int(key string, def int) int {
	str := q.ctx.Request().FormValue(key)

	// 不存在，返回默认值
	if len(str) == 0 {
		return def
	}

	// 无法转换，保存错误信息，返回默认值
	v, err := strconv.Atoi(str)
	if err != nil {
		q.errors[key] = err.Error()
		return def
	}

	return v
}

// Int64 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Queries) Int64(key string, def int64) int64 {
	str := q.ctx.Request().FormValue(key)
	if len(str) == 0 {
		return def
	}

	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		q.errors[key] = err.Error()
		return def
	}

	return v
}

// String 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Queries) String(key, def string) string {
	str := q.ctx.Request().FormValue(key)
	if len(str) == 0 {
		return def
	}
	return str
}

// Bool 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Queries) Bool(key string, def bool) bool {
	str := q.ctx.Request().FormValue(key)
	if len(str) == 0 {
		return def
	}

	v, err := strconv.ParseBool(str)
	if err != nil {
		q.errors[key] = err.Error()
		return def
	}

	return v
}

// Float64 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Queries) Float64(key string, def float64) float64 {
	str := q.ctx.Request().FormValue(key)
	if len(str) == 0 {
		return def
	}

	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		q.errors[key] = err.Error()
		return def
	}

	return v
}

// Result 返回一个 *Result 实例，若存在错误内容，
// 则这些错误内容会作为 Result.Detail 的内容一起返回。
func (q *Queries) Result(code int) *Result {
	if len(q.errors) == 0 {
		return nil
	}

	return NewResult(code, q.errors)
}

// OK 是否一切正常，若出错，则自动向 w 输出错误信息，并返回 false
func (q *Queries) OK(code int) bool {
	if len(q.errors) > 0 {
		NewResult(code, q.errors).Render(q.ctx)
		return false
	}
	return true
}
