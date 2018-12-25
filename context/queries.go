// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"strconv"

	"github.com/issue9/query"
)

// Queries 用于处理路径中的查询参数。
//  q,_ := ctx.Queries()
//  page := q.Int64("page", 1)
//  size := q.Int64("size", 20)
//  if q.HasErrors() {
//      // do something
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
	str := q.ctx.Request.FormValue(key)

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
	str := q.ctx.Request.FormValue(key)
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
	str := q.ctx.Request.FormValue(key)
	if len(str) == 0 {
		return def
	}
	return str
}

// Bool 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Queries) Bool(key string, def bool) bool {
	str := q.ctx.Request.FormValue(key)
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
	str := q.ctx.Request.FormValue(key)
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

// HasErrors 是否存在错误内容。
func (q *Queries) HasErrors() bool {
	return len(q.errors) > 0
}

// Errors 所有的错误信息
func (q *Queries) Errors() map[string]string {
	return q.errors
}

// Result 转换成 Result 对象
//
// code 是作为 Result.Code 从错误消息中查找，如果不存在，则 panic。
// Queries.errors 将会作为 Result.Detail 的内容。
func (q *Queries) Result(code int) *Result {
	return q.ctx.NewResult(code).SetDetail(q.Errors())
}

// QueryObject 将查询参数解析到一个对象中。
//
// 返回的是每一个字段对应的错误信息。
//
// 具体的文档信息可以参考 https://github.com/issue9/query
func (ctx *Context) QueryObject(v interface{}) (errors map[string]string) {
	return query.Parse(ctx.Request, v)
}
