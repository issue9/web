// SPDX-License-Identifier: MIT

package context

import (
	"net/url"
	"strconv"

	"github.com/issue9/query/v2"
)

// Queries 用于处理路径中的查询参数
//  q,_ := ctx.Queries()
//  page := q.Int64("page", 1)
//  size := q.Int64("size", 20)
//  if q.HasErrors() {
//      // do something
//      return
//  }
type Queries struct {
	ctx     *Context
	errors  ResultFields
	queries url.Values
}

// Queries 声明一个新的 Queries 实例
func (ctx *Context) Queries() (*Queries, error) {
	queries, err := url.ParseQuery(ctx.Request.URL.RawQuery)
	if err != nil {
		return nil, err
	}

	return &Queries{
		ctx:     ctx,
		errors:  ResultFields{},
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
// 若不存在则返回 def 作为其默认值。
// 若是无法转换，则会保存错误信息
func (q *Queries) Int(key string, def int) int {
	str := q.getItem(key)

	// 不存在，返回默认值
	if len(str) == 0 {
		return def
	}

	// 无法转换，保存错误信息，返回默认值
	v, err := strconv.Atoi(str)
	if err != nil {
		q.errors.Add(key, err.Error())
		return def
	}

	return v
}

// Int64 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。
func (q *Queries) Int64(key string, def int64) int64 {
	str := q.getItem(key)
	if len(str) == 0 {
		return def
	}

	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		q.errors.Add(key, err.Error())
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
// 若不存在则返回 def 作为其默认值。
func (q *Queries) Bool(key string, def bool) bool {
	str := q.getItem(key)
	if len(str) == 0 {
		return def
	}

	v, err := strconv.ParseBool(str)
	if err != nil {
		q.errors.Add(key, err.Error())
		return def
	}

	return v
}

// Float64 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。
func (q *Queries) Float64(key string, def float64) float64 {
	str := q.getItem(key)
	if len(str) == 0 {
		return def
	}

	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		q.errors.Add(key, err.Error())
		return def
	}

	return v
}

// HasErrors 是否存在错误内容
func (q *Queries) HasErrors() bool {
	return len(q.errors) > 0
}

// Errors 所有的错误信息
func (q *Queries) Errors() ResultFields {
	return q.errors
}

// Result 转换成 CTXResult 对象
//
// code 是作为 CTXResult.Code 从错误消息中查找，如果不存在，则 panic。
// Queries.errors 将会作为 CTXResult.Fields 的内容。
func (q *Queries) Result(code int) *CTXResult {
	return q.ctx.NewResultWithFields(code, q.Errors())
}

// Object 将查询参数解析到一个对象中
//
// 具体的文档信息可以参考 https://github.com/issue9/query
func (q *Queries) Object(v interface{}) {
	errors := query.Parse(q.queries, v)
	for key, vals := range errors {
		q.errors.Add(key, vals...)
	}

	if vv, ok := v.(Validator); ok {
		errors = vv.Validate(q.ctx)
		for key, vals := range errors {
			q.errors.Add(key, vals...)
		}
	}
}

// QueryObject 将查询参数解析到一个对象中
//
// 如果 err 不为 nil，表示 URL 中的查询参数格式有误；
// ok 表示是否正常解析和验证了查询参数，true 表示一切正常，false
// 表示解析或是验证出错，并以 code 作为错误代码输出到客户端。
func (ctx *Context) QueryObject(v interface{}, code int) (ok bool, err error) {
	q, err := ctx.Queries()
	if err != nil {
		return false, err
	}

	q.Object(v)

	if len(q.errors) > 0 {
		q.Result(code).Render()
		return false, nil
	}

	return true, nil
}
