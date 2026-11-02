// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/issue9/query/v3"

	"github.com/issue9/web/filter"
	"github.com/issue9/web/locales"
)

var pathPool = &sync.Pool{New: func() any { return &Paths{} }}

var queryPool = &sync.Pool{New: func() any { return &Queries{} }}

// ContextFilter 数据的验证和修正接口
//
// 通过 [Context] 从客户端提交的数据对象，如果实现了此接口，那么在解码时会调用此接口的方法验证数据。
type ContextFilter interface {
	FilterContext(*filter.Filter, *Context)
}

// Paths 提供对路径参数的处理
type Paths struct {
	filter  *filter.Filter
	ctx     *Context
	problem *Problem
}

// Queries 提供对查询参数的处理
type Queries struct {
	filter  *filter.Filter
	ctx     *Context
	problem *Problem

	queries url.Values
}

func (ctx *Context) newFilter(p *Problem, exitAtError bool) *filter.Filter {
	return filter.New("", exitAtError, func(name string, reason LocaleStringer) {
		p.WithParam(name, reason.LocaleString(ctx.LocalePrinter()))
	})
}

// Paths 声明一个用于获取路径参数的对象
//
// 返回对象的生命周期在 [Context] 结束时也随之结束。
func (ctx *Context) Paths(exitAtError bool) *Paths {
	p := newProblem()

	pp := pathPool.Get().(*Paths)
	if pp.filter == nil {
		pp.filter = ctx.newFilter(p, exitAtError)
	} else {
		pp.filter.Reset("", exitAtError, func(name string, reason LocaleStringer) {
			p.WithParam(name, reason.LocaleString(ctx.LocalePrinter()))
		})
	}
	pp.ctx = ctx
	pp.problem = p

	ctx.OnExit(func(*Context, int) { queryPool.Put(pp) })

	return pp
}

func (p *Paths) continueNext() bool { return !p.filter.ExitOnError() || len(p.problem.Params) == 0 }

func (p *Paths) Context() *Context { return p.ctx }

// ID 返回 key 所表示的值且必须大于 0
func (p *Paths) ID(key string) int64 {
	if !p.continueNext() {
		return 0
	}

	id, err := p.ctx.Route().Int(key)
	if err != nil {
		p.filter.AddError(key, err)
		return 0
	} else if id <= 0 {
		p.filter.AddReason(key, locales.ShouldGreatThan(0))
		return 0
	}
	return id
}

// Int64 获取参数 key 所代表的值并转换成 int64 类型
func (p *Paths) Int64(key string) int64 {
	if !p.continueNext() {
		return 0
	}

	ret, err := p.Context().Route().Int(key)
	if err != nil {
		p.filter.AddError(key, err)
		return 0
	}
	return ret
}

// String 获取参数 key 所代表的值并转换成 string
func (p *Paths) String(key string) string {
	if !p.continueNext() {
		return ""
	}

	ret, err := p.Context().Route().String(key)
	if err != nil {
		p.filter.AddError(key, err)
		return ""
	}
	return ret
}

// Bool 获取参数 key 所代表的值并转换成 bool
//
// 由 [strconv.ParseBool] 进行转换。
func (p *Paths) Bool(key string) bool {
	if !p.continueNext() {
		return false
	}

	ret, err := p.Context().Route().Bool(key)
	if err != nil {
		p.filter.AddError(key, err)
	}
	return ret
}

// Float64 获取参数 key 所代表的值并转换成 float64
func (p *Paths) Float64(key string) float64 {
	if !p.continueNext() {
		return 0
	}

	ret, err := p.Context().Route().Float(key)
	if err != nil {
		p.filter.AddError(key, err)
	}
	return ret
}

// Problem 如果有错误信息转换成 [Problem] 否则返回 nil
func (p *Paths) Problem(id string) Responser {
	if p.filter.HasError() {
		return p.Context().initProblem(p.problem, id)
	}
	return nil
}

// PathID 获取地址参数中表示 key 的值并转换成大于 0 的 int64
//
// NOTE: 若需要获取多个参数，使用 [Context.Paths] 会更方便。
func (ctx *Context) PathID(key, id string) (int64, Responser) {
	ret, err := ctx.Route().Int(key)
	if err != nil {
		return 0, ctx.Problem(id).WithParam(key, Phrase(err.Error()).LocaleString(ctx.LocalePrinter()))
	} else if ret <= 0 {
		return 0, ctx.Problem(id).WithParam(key, locales.ShouldGreatThan(0).LocaleString(ctx.LocalePrinter()))
	}
	return ret, nil
}

// PathInt64 取地址参数中的 key 表示的值并尝试转换成 int64 类型
//
// NOTE: 若需要获取多个参数，可以使用 [Context.Paths] 获取会更方便。
func (ctx *Context) PathInt64(key, id string) (int64, Responser) {
	ret, err := ctx.Route().Int(key)
	if err != nil {
		msg := Phrase(err.Error()).LocaleString(ctx.LocalePrinter())
		return 0, ctx.Problem(id).WithParam(key, msg)
	}
	return ret, nil
}

// PathString 取地址参数中的 key 表示的值并转换成 string 类型
//
// NOTE: 若需要获取多个参数，可以使用 [Context.Paths] 获取会更方便。
func (ctx *Context) PathString(key, id string) (string, Responser) {
	ret, err := ctx.Route().String(key)
	if err != nil {
		msg := Phrase(err.Error()).LocaleString(ctx.LocalePrinter())
		return "", ctx.Problem(id).WithParam(key, msg)
	}
	return ret, nil
}

// Queries 声明一个用于获取查询参数的对象
//
// 返回对象的生命周期在 [Context] 结束时也随之结束。
func (ctx *Context) Queries(exitAtError bool) (*Queries, error) {
	if ctx.queries != nil {
		return ctx.queries, nil
	}

	values, err := url.ParseQuery(ctx.Request().URL.RawQuery)
	if err != nil {
		return nil, err
	}

	p := newProblem()
	q := queryPool.Get().(*Queries)
	if q.filter == nil {
		q.filter = ctx.newFilter(p, exitAtError)
	} else {
		q.filter.Reset("", exitAtError, func(name string, reason LocaleStringer) {
			p.WithParam(name, reason.LocaleString(ctx.LocalePrinter()))
		})
	}
	q.queries = values
	q.ctx = ctx
	q.problem = p

	ctx.queries = q
	ctx.OnExit(func(*Context, int) { queryPool.Put(q) })
	return q, nil
}

func (q *Queries) continueNext() bool { return !q.filter.ExitOnError() || len(q.problem.Params) == 0 }

func (q *Queries) Context() *Context { return q.ctx }

func (q *Queries) getItem(key string) string { return q.queries.Get(key) }

// Int 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。若是无法转换，则会保存错误信息并返回 def。
func (q *Queries) Int(key string, def int) int {
	if !q.continueNext() {
		return 0
	}

	str := q.getItem(key)
	if len(str) == 0 { // 不存在，返回默认值
		return def
	}

	v, err := strconv.Atoi(str)
	if err != nil { // strconv.Atoi 不可能返回 LocaleStringer 接口的数据
		q.filter.AddReason(key, Phrase(err.Error()))
		return def
	}
	return v
}

// Int64 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。若是无法转换，则会保存错误信息并返回 def。
func (q *Queries) Int64(key string, def int64) int64 {
	if !q.continueNext() {
		return 0
	}

	str := q.getItem(key)
	if len(str) == 0 {
		return def
	}

	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil { // strconv.ParseInt 不可能返回 LocaleStringer 接口的数据
		q.filter.AddReason(key, Phrase(err.Error()))
		return def
	}
	return v
}

// String 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。
func (q *Queries) String(key, def string) string {
	if !q.continueNext() {
		return ""
	}

	if str := q.getItem(key); str != "" {
		return str
	}
	return def
}

// Bool 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。若是无法转换，则会保存错误信息并返回 def。
func (q *Queries) Bool(key string, def bool) bool {
	if !q.continueNext() {
		return false
	}

	str := q.getItem(key)
	if len(str) == 0 {
		return def
	}

	v, err := strconv.ParseBool(str)
	if err != nil { // strconv.ParseBool 不可能返回 LocaleStringer 接口的数据
		q.filter.AddReason(key, Phrase(err.Error()))
		return def
	}
	return v
}

// Float64 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。若是无法转换，则会保存错误信息并返回 def。
func (q *Queries) Float64(key string, def float64) float64 {
	if !q.continueNext() {
		return 0
	}

	str := q.getItem(key)
	if len(str) == 0 {
		return def
	}

	v, err := strconv.ParseFloat(str, 64)
	if err != nil { // strconv.ParseFloat 不可能返回 LocaleStringer 接口的数据
		q.filter.AddReason(key, Phrase(err.Error()))
		return def
	}
	return v
}

// Problem 如果有错误信息转换成 Problem 否则返回 nil
func (q *Queries) Problem(id string) Responser {
	if q.filter.HasError() {
		return q.Context().initProblem(q.problem, id)
	}
	return nil
}

// Object 将查询参数解析到一个对象中
//
// 具体的文档信息可以参考 [Query]。
// 如果 v 实现了 [ContextFilter] 接口，则在读取数据之后，会调用该接口方法。
//
// [Query]: https://github.com/issue9/query
func (q *Queries) Object(v any) {
	query.ParseWithLog(q.queries, v, func(field string, err error) {
		var msg LocaleStringer
		if ls, ok := err.(LocaleStringer); ok {
			msg = ls
		} else {
			msg = Phrase(err.Error())
		}

		q.filter.AddReason(field, msg)
	})

	if q.continueNext() {
		if s, ok := v.(ContextFilter); ok {
			s.FilterContext(q.filter, q.Context())
		}
	}
}

// QueryObject 将查询参数解析到一个对象中
func (ctx *Context) QueryObject(exitAtError bool, v any, id string) Responser {
	q, err := ctx.Queries(exitAtError)
	if err != nil {
		return ctx.Log(err, id)
	}
	q.Object(v)
	return q.Problem(id)
}

// RequestBody 用户提交的内容
func (ctx *Context) RequestBody() io.Reader { return ctx.requestBody }

// Unmarshal 将提交的内容解码到 v
func (ctx *Context) Unmarshal(v any) error {
	if ctx.Request().ContentLength == 0 {
		return nil
	}

	if ctx.inputMimetype == nil { // 客户端未指定 content-type，但是又有内容要输出。
		return NewLocaleError("the client miss content-type header")
	}
	return ctx.inputMimetype(ctx.RequestBody(), v)
}

// Read 从客户端读取数据并转换成 v 对象
//
// 如果 v 实现了 [ContextFilter] 接口，则在读取数据之后，会调用该接口方法。
// 如果验证失败，会返回以 id 作为错误代码的 [Problem] 对象。
func (ctx *Context) Read(exitAtError bool, v any, id string) Responser {
	if err := ctx.Unmarshal(v); err != nil {
		return ctx.Log(err, ProblemUnprocessableEntity)
	}

	p := newProblem()
	if vv, ok := v.(ContextFilter); ok {
		vv.FilterContext(ctx.newFilter(p, exitAtError), ctx)
		return ctx.initProblem(p, id)
	}
	return nil
}

// Request 返回原始的请求对象
func (ctx *Context) Request() *http.Request { return ctx.request }
