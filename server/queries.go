// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/issue9/query/v3"

	"github.com/issue9/web/problem"
)

// Queries 用于处理路径中的查询参数
//
//  q,_ := ctx.Queries()
//  page := q.Int64("page", 1)
//  size := q.Int64("size", 20)
//  if q.HasErrors() {
//      // do something
//      return
//  }
type Queries struct {
	ctx     *Context
	fields  FieldErrs
	queries url.Values
}

// Queries 声明一个新的 Queries 实例
func (ctx *Context) Queries() (*Queries, error) {
	queries, err := url.ParseQuery(ctx.Request().URL.RawQuery)
	if err != nil {
		return nil, err
	}

	return &Queries{
		ctx:     ctx,
		fields:  FieldErrs{},
		queries: queries,
	}, nil
}

func (q *Queries) getItem(key string) (val string) {
	if v, found := q.queries[key]; found {
		val = v[0]
	}
	return
}

// Int 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。若是无法转换，则会保存错误信息并返回 def。
func (q *Queries) Int(key string, def int) int {
	str := q.getItem(key)

	// 不存在，返回默认值
	if len(str) == 0 {
		return def
	}

	// 无法转换，保存错误信息，返回默认值
	v, err := strconv.Atoi(str)
	if err != nil {
		q.fields.Add(key, err.Error())
		return def
	}

	return v
}

// Int64 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。若是无法转换，则会保存错误信息并返回 def。
func (q *Queries) Int64(key string, def int64) int64 {
	str := q.getItem(key)
	if len(str) == 0 {
		return def
	}

	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		q.fields.Add(key, err.Error())
		return def
	}

	return v
}

// String 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。
func (q *Queries) String(key, def string) string {
	str := q.getItem(key)
	if len(str) == 0 {
		return def
	}
	return str
}

// Bool 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。若是无法转换，则会保存错误信息并返回 def。
func (q *Queries) Bool(key string, def bool) bool {
	str := q.getItem(key)
	if len(str) == 0 {
		return def
	}

	v, err := strconv.ParseBool(str)
	if err != nil {
		q.fields.Add(key, err.Error())
		return def
	}

	return v
}

// Float64 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。若是无法转换，则会保存错误信息并返回 def。
func (q *Queries) Float64(key string, def float64) float64 {
	str := q.getItem(key)
	if len(str) == 0 {
		return def
	}

	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		q.fields.Add(key, err.Error())
		return def
	}

	return v
}

// HasErrors 是否存在错误内容
func (q *Queries) HasErrors() bool { return len(q.fields) > 0 }

// Errors 所有的错误信息
func (q *Queries) Errors() FieldErrs { return q.fields }

// Result 转换成 Response 对象
func (q *Queries) Result(id string) Responser {
	if q.HasErrors() {
		return q.ctx.Problem(problem.NewInvalidParamsProblem(q.Errors()), id)
	}
	return nil
}

// Object 将查询参数解析到一个对象中
//
// 具体的文档信息可以参考 https://github.com/issue9/query
//
// 如果 v 实现了 CTXSanitizer 接口，则在读取数据之后，会调用其接口函数。
// 如果验证失败，错误信息存入 q.errors。
func (q *Queries) Object(v any) {
	errs := query.Parse(q.queries, v)
	for k, err := range errs {
		q.fields.Add(k, err.Error())
	}

	if vv, ok := v.(CTXSanitizer); ok {
		q.fields.Merge(vv.CTXSanitize(q.ctx))
	}
}

// QueryObject 将查询参数解析到一个对象中
func (ctx *Context) QueryObject(v any, code string) Responser {
	q, err := ctx.Queries()
	if err != nil {
		return ctx.Error(http.StatusUnprocessableEntity, err)
	}

	q.Object(v)
	return q.Result(code)
}
