// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/issue9/localeutil"
	"github.com/issue9/query/v3"

	"github.com/issue9/web/problem"
	"github.com/issue9/web/server/response"
)

// Queries 用于处理路径中的查询参数
//
//  q,_ := ctx.Queries()
//  page := q.Int64("page", 1)
//  size := q.Int64("size", 20)
type Queries struct {
	ctx     *Context
	v       *problem.Validation
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
		v:       ctx.NewValidation(),
		queries: queries,
	}, nil
}

func (q *Queries) getItem(key string) (val string) { return q.queries.Get(key) }

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
		q.v.Add(key, localeutil.Phrase(err.Error()))
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
		q.v.Add(key, localeutil.Phrase(err.Error()))
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
		q.v.Add(key, localeutil.Phrase(err.Error()))
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
		q.v.Add(key, localeutil.Phrase(err.Error()))
		return def
	}

	return v
}

// Problem 转换成 Responser 对象
func (q *Queries) Problem(id string) response.Responser {
	return q.v.Problem(id, q.ctx.LocalePrinter())
}

// Object 将查询参数解析到一个对象中
//
// 具体的文档信息可以参考 https://github.com/issue9/query
//
// 如果 v 实现了 CTXSanitizer 接口，则在读取数据之后，会调用其接口函数。
func (q *Queries) Object(v any, id string) response.Responser {
	errs := query.Parse(q.queries, v)
	for k, err := range errs {
		q.v.Add(k, localeutil.Phrase(err.Error()))
	}
	if len(errs) > 0 {
		return q.Problem(id)
	}

	if vv, ok := v.(CTXSanitizer); ok {
		return vv.CTXSanitize(q.ctx)
	}
	return nil
}

// QueryObject 将查询参数解析到一个对象中
func (ctx *Context) QueryObject(v any, id string) response.Responser {
	q, err := ctx.Queries()
	if err != nil {
		return ctx.Error(http.StatusUnprocessableEntity, err)
	}

	return q.Object(v, id)
}
