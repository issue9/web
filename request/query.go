// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package request

import (
	"net/http"

	"github.com/issue9/web/result"
	"github.com/issue9/web/types"
)

// Query 用于处理路径中的查询参数。用法类似于 flag
//  q,_ := NewQuery(r, false)
//  page := q.Int64("page", 1)
//  size := q.Int64("size", 20)
//  if r := q.Result();r.HasDetail(){
//      rslt.RenderJSON(w)
//      return
//  }
type Query struct {
	errors  map[string]string
	request *http.Request
}

// NewQuery 声明一个新的 Query 实例
func NewQuery(r *http.Request) *Query {
	return &Query{
		errors:  map[string]string{},
		request: r,
	}
}

func (q *Query) parseOne(key string, val value) {
	v := q.request.FormValue(key)

	if len(v) == 0 { // 不存在，使用原来的值
		return
	}

	if err := val.set(v); err != nil {
		q.errors[key] = err.Error()
	}
}

// Int 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
//
// 若是无法转换，则会保存错误信息
func (q *Query) Int(key string, def int) int {
	i := new(int)
	q.intVar(i, key, def)
	return *i
}

// IntVar 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Query) IntVar(i *int, key string, def int) *Query {
	q.intVar(i, key, def)
	return q
}

func (q *Query) intVar(i *int, key string, def int) {
	*i = def
	q.parseOne(key, (*intValue)(i))
}

// Int64 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Query) Int64(key string, def int64) int64 {
	i := new(int64)
	q.int64Var(i, key, def)
	return *i
}

// Int64Var 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Query) Int64Var(i *int64, key string, def int64) *Query {
	q.int64Var(i, key, def)
	return q
}

func (q *Query) int64Var(i *int64, key string, def int64) {
	*i = def
	q.parseOne(key, (*int64Value)(i))
}

// String 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Query) String(key, def string) string {
	i := new(string)
	q.stringVar(i, key, def)
	return *i
}

// StringVar 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Query) StringVar(i *string, key string, def string) *Query {
	q.stringVar(i, key, def)
	return q
}

func (q *Query) stringVar(i *string, key string, def string) {
	*i = def
	q.parseOne(key, (*stringValue)(i))
}

// Bool 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Query) Bool(key string, def bool) bool {
	i := new(bool)
	q.boolVar(i, key, def)
	return *i
}

// BoolVar 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Query) BoolVar(i *bool, key string, def bool) *Query {
	q.boolVar(i, key, def)
	return q
}

func (q *Query) boolVar(i *bool, key string, def bool) {
	*i = def
	q.parseOne(key, (*boolValue)(i))
}

// Float64 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Query) Float64(key string, def float64) float64 {
	i := new(float64)
	q.float64Var(i, key, def)
	return *i
}

// Float64Var 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Query) Float64Var(i *float64, key string, def float64) *Query {
	q.float64Var(i, key, def)
	return q
}

func (q *Query) float64Var(i *float64, key string, def float64) {
	*i = def
	q.parseOne(key, (*float64Value)(i))
}

// Result 返回一个 *result.Result 实例，若存在错误内容，
// 则这些错误内容会作为 result.Result.Detail 的内容一起返回。
func (q *Query) Result(code int) *result.Result {
	if len(q.errors) == 0 {
		return result.New(code)
	}

	return result.NewWithDetail(code, q.errors)
}

// OK 是否一切正常，若出错，则自动向 w 输出错误信息，并返回 false
func (q *Query) OK(ctx types.Context, code int) bool {
	if len(q.errors) > 0 {
		q.Result(code).Render(ctx)
		return false
	}
	return true
}
