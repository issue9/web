// SPDX-License-Identifier: MIT

package web

import (
	"errors"
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

const defaultBodyBufferSize = 256

var queryPool = &sync.Pool{New: func() any { return &Queries{} }}

// Paths 提供对路径参数的处理
type Paths struct {
	filter *FilterProblem
}

// Queries 提供对查询参数的处理
type Queries struct {
	filter  *FilterProblem
	queries url.Values
}

// CTXFilter 在 [Context] 关联的上下文环境中对用户提交的数据进行验证和修正
type CTXFilter interface {
	CTXFilter(*FilterProblem)
}

// Paths 声明一个用于获取路径参数的对象
//
// 返回对象的生命周期在 [Context] 结束时也随之结束。
func (ctx *Context) Paths(exitAtError bool) *Paths {
	return &Paths{filter: ctx.NewFilterProblem(exitAtError)}
}

// ID 返回 key 所表示的值且必须大于 0
func (p *Paths) ID(key string) int64 {
	if !p.filter.continueNext() {
		return 0
	}

	id, err := p.filter.Context().Route().Params().Int(key)
	if err != nil {
		p.filter.AddError(key, err)
		return 0
	} else if id <= 0 {
		p.filter.Add(key, locales.ShouldGreatThanZero)
		return 0
	}
	return id
}

// Int64 获取参数 key 所代表的值并转换成 int64 类型
func (p *Paths) Int64(key string) int64 {
	if !p.filter.continueNext() {
		return 0
	}

	ret, err := p.filter.Context().Route().Params().Int(key)
	if err != nil {
		p.filter.AddError(key, err)
		return 0
	}
	return ret
}

// String 获取参数 key 所代表的值并转换成 string
func (p *Paths) String(key string) string {
	if !p.filter.continueNext() {
		return ""
	}

	ret, err := p.filter.Context().Route().Params().String(key)
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
	if !p.filter.continueNext() {
		return false
	}

	ret, err := p.filter.Context().Route().Params().Bool(key)
	if err != nil {
		p.filter.AddError(key, err)
	}
	return ret
}

// Float64 获取参数 key 所代表的值并转换成 float64
func (p *Paths) Float64(key string) float64 {
	if !p.filter.continueNext() {
		return 0
	}

	ret, err := p.filter.Context().Route().Params().Float(key)
	if err != nil {
		p.filter.AddError(key, err)
	}
	return ret
}

// Problem 如果有错误信息转换成 [Problem] 否则返回 nil
func (p *Paths) Problem(id string) Responser { return p.filter.Problem(id) }

// PathID 获取地址参数中表示 key 的值并转换成大于 0 的 int64
//
// NOTE: 若需要获取多个参数，使用 [Context.Paths] 会更方便。
func (ctx *Context) PathID(key, id string) (int64, Responser) {
	p := ctx.LocalePrinter()
	ret, err := ctx.Route().Params().Int(key)
	if err != nil {
		return 0, ctx.Problem(id).WithParam(key, Phrase(err.Error()).LocaleString(p))
	} else if ret <= 0 {
		return 0, ctx.Problem(id).WithParam(key, locales.ShouldGreatThanZero.LocaleString(p))
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

// PathString 取地址参数中的 key 表示的值并尝试转换成 string 类型
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
	queries, err := url.ParseQuery(ctx.Request().URL.RawQuery)
	if err != nil {
		return nil, err
	}

	q := queryPool.Get().(*Queries)
	q.filter = ctx.NewFilterProblem(exitAtError)
	q.queries = queries
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
		q.filter.Add(key, Phrase(err.Error()))
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
		q.filter.Add(key, Phrase(err.Error()))
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
		q.filter.Add(key, Phrase(err.Error()))
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
		q.filter.Add(key, Phrase(err.Error()))
		return def
	}
	return v
}

// Problem 如果有错误信息转换成 Problem 否则返回 nil
func (q *Queries) Problem(id string) Responser { return q.filter.Problem(id) }

// Object 将查询参数解析到一个对象中
//
// 具体的文档信息可以参考 [Query]。
// 如果 v 实现了 [CTXFilter] 接口，则在读取数据之后，会调用该接口方法。
//
// [Query]: https://github.com/issue9/query
func (q *Queries) Object(v any, id string) {
	query.ParseWithLog(q.queries, v, func(field string, err error) {
		var msg LocaleStringer
		if ls, ok := err.(LocaleStringer); ok {
			msg = ls
		} else {
			msg = Phrase(err.Error())
		}

		q.filter.Add(field, msg)
	})

	if q.filter.continueNext() {
		if s, ok := v.(CTXFilter); ok {
			s.CTXFilter(q.filter)
		}
	}
}

// QueryObject 将查询参数解析到一个对象中
func (ctx *Context) QueryObject(exitAtError bool, v any, id string) Responser {
	q, err := ctx.Queries(exitAtError)
	if err != nil {
		return ctx.Error(err, id)
	}
	q.Object(v, id)

	return q.Problem(id)
}

// RequestBody 获取用户提交的内容
//
// 此方法可多次调用且返回正确内容，但是不能对返回内容作直接修改。不存在 body 时，返回 nil。
func (ctx *Context) RequestBody() (body []byte, err error) {
	if ctx.read {
		return ctx.requestBody, nil
	}
	req := ctx.Request()
	if req.ContentLength == 0 {
		ctx.read = true
		return nil, nil
	}

	var reader io.Reader = req.Body
	if !header.CharsetIsNop(ctx.inputCharset) {
		reader = transform.NewReader(reader, ctx.inputCharset.NewDecoder())
	}

	if ctx.requestBody == nil {
		size := req.ContentLength
		if size == -1 {
			size = defaultBodyBufferSize
		}
		ctx.requestBody = make([]byte, 0, size)
	}

	for {
		if len(ctx.requestBody) == cap(ctx.requestBody) {
			ctx.requestBody = append(ctx.requestBody, 0)[:len(ctx.requestBody)]
		}
		n, err := reader.Read(ctx.requestBody[len(ctx.requestBody):cap(ctx.requestBody)])
		ctx.requestBody = ctx.requestBody[:len(ctx.requestBody)+n]
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
	}

	ctx.read = true
	return ctx.requestBody, err
}

// Unmarshal 将提交的内容转换成 v 对象
func (ctx *Context) Unmarshal(v any) error {
	body, err := ctx.RequestBody()
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return nil
	}
	if ctx.inputMimetype == nil {
		return NewLocaleError("the client did not specify content-type header")
	}
	return ctx.inputMimetype(body, v)
}

// Read 从客户端读取数据并转换成 v 对象
//
// 如果 v 实现了 [CTXFilter] 接口，则在读取数据之后，会调用该接口方法。
// 如果验证失败，会返回以 id 作为错误代码的 [Problem] 对象。
func (ctx *Context) Read(exitAtError bool, v any, id string) Responser {
	if err := ctx.Unmarshal(v); err != nil {
		return ctx.Error(err, ProblemUnprocessableEntity)
	}

	if vv, ok := v.(CTXFilter); ok {
		va := ctx.NewFilterProblem(exitAtError)
		vv.CTXFilter(va)
		return va.Problem(id)
	}
	return nil
}

// Request 返回原始的请求对象
func (ctx *Context) Request() *http.Request { return ctx.request }
