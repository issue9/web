// SPDX-License-Identifier: MIT

package web

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/issue9/query/v3"
	"golang.org/x/text/transform"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/locales"
)

var queryPool = &sync.Pool{New: func() any { return &Queries{} }}

// Paths 提供对路径参数的处理
type Paths FilterContext

// Queries 提供对查询参数的处理
type Queries struct {
	filter  *FilterContext
	queries url.Values
}

// Paths 声明一个用于获取路径参数的对象
//
// 返回对象的生命周期在 [Context] 结束时也随之结束。
func (ctx *Context) Paths(exitAtError bool) *Paths {
	return (*Paths)(ctx.newFilterContext(exitAtError))
}

func (p *Paths) filter() *FilterContext { return (*FilterContext)(p) }

// ID 返回 key 所表示的值且必须大于 0
func (p *Paths) ID(key string) int64 {
	if !p.filter().continueNext() {
		return 0
	}

	id, err := p.filter().Context().Route().Params().Int(key)
	if err != nil {
		p.filter().AddError(key, err)
		return 0
	} else if id <= 0 {
		p.filter().AddReason(key, locales.ShouldGreatThan(0))
		return 0
	}
	return id
}

// Int64 获取参数 key 所代表的值并转换成 int64 类型
func (p *Paths) Int64(key string) int64 {
	if !p.filter().continueNext() {
		return 0
	}

	ret, err := p.filter().Context().Route().Params().Int(key)
	if err != nil {
		p.filter().AddError(key, err)
		return 0
	}
	return ret
}

// String 获取参数 key 所代表的值并转换成 string
func (p *Paths) String(key string) string {
	if !p.filter().continueNext() {
		return ""
	}

	ret, err := p.filter().Context().Route().Params().String(key)
	if err != nil {
		p.filter().AddError(key, err)
		return ""
	}
	return ret
}

// Bool 获取参数 key 所代表的值并转换成 bool
//
// 由 [strconv.ParseBool] 进行转换。
func (p *Paths) Bool(key string) bool {
	if !p.filter().continueNext() {
		return false
	}

	ret, err := p.filter().Context().Route().Params().Bool(key)
	if err != nil {
		p.filter().AddError(key, err)
	}
	return ret
}

// Float64 获取参数 key 所代表的值并转换成 float64
func (p *Paths) Float64(key string) float64 {
	if !p.filter().continueNext() {
		return 0
	}

	ret, err := p.filter().Context().Route().Params().Float(key)
	if err != nil {
		p.filter().AddError(key, err)
	}
	return ret
}

// Problem 如果有错误信息转换成 [Problem] 否则返回 nil
func (p *Paths) Problem(id string) Responser { return p.filter().Problem(id) }

// PathID 获取地址参数中表示 key 的值并转换成大于 0 的 int64
//
// NOTE: 若需要获取多个参数，使用 [Context.Paths] 会更方便。
func (ctx *Context) PathID(key, id string) (int64, Responser) {
	ret, err := ctx.Route().Params().Int(key)
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
	ret, err := ctx.Route().Params().Int(key)
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
	ret, err := ctx.Route().Params().String(key)
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

	q := queryPool.Get().(*Queries)
	q.filter = ctx.newFilterContext(exitAtError)
	q.queries = values
	ctx.queries = q
	ctx.OnExit(func(*Context, int) { queryPool.Put(q) })
	return q, nil
}

func (q *Queries) getItem(key string) string { return q.queries.Get(key) }

// Int 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。若是无法转换，则会保存错误信息并返回 def。
func (q *Queries) Int(key string, def int) int {
	if !q.filter.continueNext() {
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
	if !q.filter.continueNext() {
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
	if !q.filter.continueNext() {
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
	if !q.filter.continueNext() {
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
	if !q.filter.continueNext() {
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
func (q *Queries) Problem(id string) Responser { return q.filter.Problem(id) }

// Object 将查询参数解析到一个对象中
//
// 具体的文档信息可以参考 [Query]。
// 如果 v 实现了 [Filter] 接口，则在读取数据之后，会调用该接口方法。
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

	if q.filter.continueNext() {
		if s, ok := v.(Filter); ok {
			s.Filter(q.filter)
		}
	}
}

// QueryObject 将查询参数解析到一个对象中
func (ctx *Context) QueryObject(exitAtError bool, v any, id string) Responser {
	q, err := ctx.Queries(exitAtError)
	if err != nil {
		return ctx.Error(err, id)
	}
	q.Object(v)
	return q.Problem(id)
}

// RequestBody 用户提交的内容
func (ctx *Context) RequestBody() io.Reader {
	r := ctx.Request().Body // 作为服务端使用，Body 始终不为空，且不需要调用 Close

	if !header.CharsetIsNop(ctx.inputCharset) {
		return transform.NewReader(r, ctx.inputCharset.NewDecoder())
	}
	return r
}

// Unmarshal 将提交的内容转换成 v 对象
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
// 如果 v 实现了 [Filter] 接口，则在读取数据之后，会调用该接口方法。
// 如果验证失败，会返回以 id 作为错误代码的 [Problem] 对象。
func (ctx *Context) Read(exitAtError bool, v any, id string) Responser {
	if err := ctx.Unmarshal(v); err != nil {
		return ctx.Error(err, ProblemUnprocessableEntity)
	}

	if vv, ok := v.(Filter); ok {
		f := ctx.newFilterContext(exitAtError)
		vv.Filter(f)
		return f.Problem(id)
	}
	return nil
}

// Request 返回原始的请求对象
func (ctx *Context) Request() *http.Request { return ctx.request }
