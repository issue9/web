// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

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

func (q *Queries) parseOne(key string, val value) {
	v := q.ctx.Request().FormValue(key)

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
func (q *Queries) Int(key string, def int) int {
	i := new(int)
	q.intVar(i, key, def)
	return *i
}

// IntVar 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Queries) IntVar(i *int, key string, def int) *Queries {
	q.intVar(i, key, def)
	return q
}

func (q *Queries) intVar(i *int, key string, def int) {
	*i = def
	q.parseOne(key, (*intValue)(i))
}

// Int64 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Queries) Int64(key string, def int64) int64 {
	i := new(int64)
	q.int64Var(i, key, def)
	return *i
}

// Int64Var 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Queries) Int64Var(i *int64, key string, def int64) *Queries {
	q.int64Var(i, key, def)
	return q
}

func (q *Queries) int64Var(i *int64, key string, def int64) {
	*i = def
	q.parseOne(key, (*int64Value)(i))
}

// String 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Queries) String(key, def string) string {
	i := new(string)
	q.stringVar(i, key, def)
	return *i
}

// StringVar 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Queries) StringVar(i *string, key string, def string) *Queries {
	q.stringVar(i, key, def)
	return q
}

func (q *Queries) stringVar(i *string, key string, def string) {
	*i = def
	q.parseOne(key, (*stringValue)(i))
}

// Bool 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Queries) Bool(key string, def bool) bool {
	i := new(bool)
	q.boolVar(i, key, def)
	return *i
}

// BoolVar 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Queries) BoolVar(i *bool, key string, def bool) *Queries {
	q.boolVar(i, key, def)
	return q
}

func (q *Queries) boolVar(i *bool, key string, def bool) {
	*i = def
	q.parseOne(key, (*boolValue)(i))
}

// Float64 从查询参数中获取指定名称的值，若不存在则返回 def 作为其默认值。
func (q *Queries) Float64(key string, def float64) float64 {
	i := new(float64)
	q.float64Var(i, key, def)
	return *i
}

// Float64Var 从查询参数中获取指定名称的值到 i，若不存在则使用 def 作为其默认值。
func (q *Queries) Float64Var(i *float64, key string, def float64) *Queries {
	q.float64Var(i, key, def)
	return q
}

func (q *Queries) float64Var(i *float64, key string, def float64) {
	*i = def
	q.parseOne(key, (*float64Value)(i))
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
		q.Result(code).Render(q.ctx)
		return false
	}
	return true
}
