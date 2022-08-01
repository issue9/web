// SPDX-License-Identifier: MIT

package server

import (
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/issue9/localeutil"
	"github.com/issue9/query/v3"
	"golang.org/x/text/transform"

	"github.com/issue9/web/problem"
	"github.com/issue9/web/server/response"
)

var tGreatThanZero = localeutil.Phrase("should great than 0")

type (
	// Params 用于处理路径中包含的参数
	//  p := ctx.Params()
	//  aid := p.Int64("aid")
	//  bid := p.Int64("bid")
	Params struct {
		ctx *Context
		v   *problem.Validation
	}

	// Queries 用于处理路径中的查询参数
	//
	//  q,_ := ctx.Queries()
	//  page := q.Int64("page", 1)
	//  size := q.Int64("size", 20)
	Queries struct {
		ctx     *Context
		v       *problem.Validation
		queries url.Values
	}
)

// Params 声明一个新的 Params 实例
func (ctx *Context) Params() *Params {
	return &Params{
		ctx: ctx,
		v:   ctx.NewValidation(),
	}
}

// ID 获取参数 key 所代表的值并转换成 int64
//
// 值必须大于 0，否则会输出错误信息，并返回零值。
func (p *Params) ID(key string) int64 {
	id, err := p.ctx.route.Params().Int(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
	} else if id <= 0 {
		p.v.Add(key, tGreatThanZero)
	}
	return id
}

// Int64 获取参数 key 所代表的值，并转换成 int64
func (p *Params) Int64(key string) int64 {
	ret, err := p.ctx.route.Params().Int(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
	}
	return ret
}

// String 获取参数 key 所代表的值并转换成 string
func (p *Params) String(key string) string {
	ret, err := p.ctx.route.Params().String(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
	}
	return ret
}

// Bool 获取参数 key 所代表的值并转换成 bool
//
// 最终会调用 strconv.ParseBool 进行转换，
// 也只有该方法中允许的字符串会被正确转换。
func (p *Params) Bool(key string) bool {
	ret, err := p.ctx.route.Params().Bool(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
	}
	return ret
}

// Float64 获取参数 key 所代表的值并转换成 float64
func (p *Params) Float64(key string) float64 {
	ret, err := p.ctx.route.Params().Float(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
	}
	return ret
}

func (p *Params) Problem(id string) response.Responser {
	return p.v.Problem(p.ctx.Server().Problems(), id, p.ctx.LocalePrinter())
}

// ParamID 获取地址参数中表示 key 的值并并转换成大于 0 的 int64
//
// NOTE: 若需要获取多个参数，使用 Context.Params 会更方便。
func (ctx *Context) ParamID(key, id string) (int64, response.Responser) {
	p := ctx.Params()
	num := p.ID(key)
	if pp := p.Problem(id); pp != nil {
		return 0, pp
	}
	return num, nil
}

// ParamInt64 取地址参数中的 key 表示的值 int64 类型值
//
// NOTE: 若需要获取多个参数，可以使用 Context.Params 获取会更方便。
func (ctx *Context) ParamInt64(key, id string) (int64, response.Responser) {
	p := ctx.Params()
	num := p.Int64(key)
	if pp := p.Problem(id); pp != nil {
		return 0, pp
	}
	return num, nil
}

// ParamString 取地址参数中的 key 表示的 string 类型值
//
// NOTE: 若需要获取多个参数，可以使用 Context.Params 获取会更方便。
func (ctx *Context) ParamString(key, id string) (string, response.Responser) {
	p := ctx.Params()
	s := p.String(key)
	if pp := p.Problem(id); pp != nil {
		return "", pp
	}
	return s, nil
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
//
// 如果没有错误，则返回 nil。
func (q *Queries) Problem(id string) response.Responser {
	return q.v.Problem(q.ctx.Server().Problems(), id, q.ctx.LocalePrinter())
}

// Object 将查询参数解析到一个对象中
//
// 具体的文档信息可以参考 https://github.com/issue9/query
//
// 如果 v 实现了 CTXSanitizer 接口，则在读取数据之后，会调用其接口函数。
func (q *Queries) Object(v any, id string) {
	for k, err := range query.Parse(q.queries, v) {
		q.v.Add(k, localeutil.Phrase(err.Error()))
	}

	if vv, ok := v.(CTXSanitizer); ok {
		if va := vv.CTXSanitize(q.ctx); va != nil {
			va.Visit(func(name string, reason localeutil.LocaleStringer) bool {
				q.v.Add(name, reason)
				return true
			})
		}
	}
}

// QueryObject 将查询参数解析到一个对象中
func (ctx *Context) QueryObject(v any, id string) response.Responser {
	q, err := ctx.Queries()
	if err != nil {
		return ctx.Error(http.StatusUnprocessableEntity, err)
	}
	q.Object(v, id)

	return q.Problem(id)
}

// Body 获取用户提交的内容
//
// 相对于 ctx.Request().Body，此函数可多次读取。不存在 body 时，返回 nil
func (ctx *Context) Body() (body []byte, err error) {
	if ctx.read {
		return ctx.body, nil
	}

	var reader io.Reader = ctx.Request().Body
	if !charsetIsNop(ctx.inputCharset) {
		reader = transform.NewReader(reader, ctx.inputCharset.NewDecoder())
	}

	if ctx.body == nil {
		ctx.body = make([]byte, 0, defaultBodyBufferSize)
	}

	for {
		if len(ctx.body) == cap(ctx.body) {
			ctx.body = append(ctx.body, 0)[:len(ctx.body)]
		}
		n, err := reader.Read(ctx.body[len(ctx.body):cap(ctx.body)])
		ctx.body = ctx.body[:len(ctx.body)+n]
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}

	ctx.read = true
	return ctx.body, err
}

// Unmarshal 将提交的内容转换成 v 对象
func (ctx *Context) Unmarshal(v any) error {
	body, err := ctx.Body()
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return nil
	}
	return ctx.inputMimetype(body, v)
}

// Read 从客户端读取数据并转换成 v 对象
//
// 功能与 Unmarshal() 相同，只不过 Read() 在出错时，返回的不是 error，
// 而是一个表示错误信息的 Response 对象。
//
// 如果 v 实现了 CTXSanitizer 接口，则在读取数据之后，会调用其接口函数。
// 如果验证失败，会输出以 id 作为错误代码的 Response 对象。
func (ctx *Context) Read(v any, id string) response.Responser {
	if err := ctx.Unmarshal(v); err != nil {
		return ctx.Error(http.StatusUnprocessableEntity, err)
	}

	if vv, ok := v.(CTXSanitizer); ok {
		if va := vv.CTXSanitize(ctx); va != nil {
			return va.Problem(ctx.Server().Problems(), id, ctx.LocalePrinter())
		}
	}

	return nil
}
