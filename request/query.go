// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package request

import "net/http"

// Query 用于处理路径中的查询参数。用法类似于 flag
//  q,_ := NewQuery(r, false)
//  page := q.Int64("page", 1)
//  size := q.Int64("size", 20)
//  if msg := q.Parse();len(msg)>0{
//      rslt := web.NewResultWithDetail(400, msg)
//      rslt.RenderJSON(w,r,nil)
//      return
//  }
type Query struct {
	abortOnError bool
	errors       map[string]string
	values       map[string]value
	request      *http.Request
}

// NewQuery 声明一个新的 Query 实例
func NewQuery(r *http.Request, abortOnError bool) *Query {
	return &Query{
		abortOnError: abortOnError,
		errors:       map[string]string{},
		values:       make(map[string]value, 5),
		request:      r,
	}
}

func (q *Query) parseOne(key string, val value) (ok bool) {
	v := q.request.FormValue(key)

	if len(v) == 0 { // 不存在，使用原来的值
		return true
	}

	if err := val.set(v); err != nil {
		q.errors[key] = err.Error()
		return false
	}
	return true
}

// Int 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Query) Int(key string, def int) *int {
	i := new(int)
	q.IntVar(i, key, def)
	return i
}

// IntVar 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Query) IntVar(i *int, key string, def int) {
	*i = def
	q.values[key] = (*intValue)(i)
}

// Int64 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Query) Int64(key string, def int64) *int64 {
	i := new(int64)
	q.Int64Var(i, key, def)
	return i
}

// Int64Var 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Query) Int64Var(i *int64, key string, def int64) {
	*i = def
	q.values[key] = (*int64Value)(i)
}

// ID 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。值必须大于 0
func (q *Query) ID(key string, def int64) *int64 {
	i := new(int64)
	q.IDVar(i, key, def)
	return i
}

// IDVar 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。值必须大于 0
func (q *Query) IDVar(i *int64, key string, def int64) {
	*i = def
	q.values[key] = (*int64Value)(i)
}

// String 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Query) String(key, def string) *string {
	i := new(string)
	q.StringVar(i, key, def)
	return i
}

// StringVar 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Query) StringVar(i *string, key string, def string) {
	*i = def
	q.values[key] = (*stringValue)(i)
}

// Bool 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Query) Bool(key string, def bool) *bool {
	i := new(bool)
	q.BoolVar(i, key, def)
	return i
}

// BoolVar 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Query) BoolVar(i *bool, key string, def bool) {
	*i = def
	q.values[key] = (*boolValue)(i)
}

// Float64 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Query) Float64(key string, def float64) *float64 {
	i := new(float64)
	q.Float64Var(i, key, def)
	return i
}

// Float64Var 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Query) Float64Var(i *float64, key string, def float64) {
	*i = def
	q.values[key] = (*float64Value)(i)
}

// Parse 开始解析数据，若存在错误则返回每个参数对应的错误信息
func (q *Query) Parse() map[string]string {
	for k, v := range q.values {
		ok := q.parseOne(k, v)
		if !ok && q.abortOnError {
			break
		}
	}

	return q.errors
}
