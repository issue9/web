// SPDX-License-Identifier: MIT

package server

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/issue9/localeutil"
	"github.com/issue9/query/v3"
	"golang.org/x/text/transform"

	"github.com/issue9/web/internal/header"
)

var tGreatThanZero = localeutil.Phrase("should great than 0")

var (
	paramPool = &sync.Pool{New: func() any { return &Params{} }}
	queryPool = &sync.Pool{New: func() any { return &Queries{} }}
)

type (
	Params struct {
		ctx *Context
		v   *CTXValidation
	}

	Queries struct {
		ctx     *Context
		v       *CTXValidation
		queries url.Values
	}
)

// Params 声明一个用于获取路径参数的对象
func (ctx *Context) Params() *Params {
	ps := paramPool.Get().(*Params)
	ps.ctx = ctx
	ps.v = ctx.NewValidation()
	ctx.OnExit(func(i int) {
		paramPool.Put(ps)
	})
	return ps
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

// Problem 将当前对象转换成 [Problem] 对象
//
// 仅在处理参数时有错误的情况下，才会转换成 [Problem] 对象，否则将返回空值。
func (p *Params) Problem(id string) Responser { return p.v.Problem(id) }

// ParamID 获取地址参数中表示 key 的值并并转换成大于 0 的 int64
//
// NOTE: 若需要获取多个参数，使用 [Context.Params] 会更方便。
func (ctx *Context) ParamID(key, id string) (int64, Responser) {
	p := ctx.LocalePrinter()
	ps := ctx.Server().Problems()
	ret, err := ctx.route.Params().Int(key)
	if err != nil {
		return 0, ps.Problem(id).AddParam(key, localeutil.Phrase(err.Error()).LocaleString(p))
	} else if ret <= 0 {
		return 0, ps.Problem(id).AddParam(key, tGreatThanZero.LocaleString(p))
	}
	return ret, nil
}

// ParamInt64 取地址参数中的 key 表示的值 int64 类型值
//
// NOTE: 若需要获取多个参数，可以使用 [Context.Params] 获取会更方便。
func (ctx *Context) ParamInt64(key, id string) (int64, Responser) {
	ret, err := ctx.route.Params().Int(key)
	if err != nil {
		p := ctx.LocalePrinter()
		ps := ctx.Server().Problems()
		return 0, ps.Problem(id).AddParam(key, localeutil.Phrase(err.Error()).LocaleString(p))
	}
	return ret, nil
}

// ParamString 取地址参数中的 key 表示的 string 类型值
//
// NOTE: 若需要获取多个参数，可以使用 [Context.Params] 获取会更方便。
func (ctx *Context) ParamString(key, id string) (string, Responser) {
	ret, err := ctx.route.Params().String(key)
	if err != nil {
		p := ctx.LocalePrinter()
		ps := ctx.Server().Problems()
		return "", ps.Problem(id).AddParam(key, localeutil.Phrase(err.Error()).LocaleString(p))
	}
	return ret, nil
}

// Queries 声明一个用于获取查询参数的对象
func (ctx *Context) Queries() (*Queries, error) {
	queries, err := url.ParseQuery(ctx.Request().URL.RawQuery)
	if err != nil {
		return nil, err
	}

	q := queryPool.Get().(*Queries)
	q.ctx = ctx
	q.v = ctx.NewValidation()
	q.queries = queries
	ctx.OnExit(func(i int) {
		queryPool.Put(q)
	})
	return q, nil
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

// Problem 将当前对象转换成 [Problem] 对象
//
// 仅在处理参数时有错误的情况下，才会转换成 [Problem] 对象，否则将返回空值。
func (q *Queries) Problem(id string) Responser { return q.v.Problem(id) }

// Object 将查询参数解析到一个对象中
//
// 具体的文档信息可以参考 https://github.com/issue9/query
//
// 如果 v 实现了 [CTXSanitizer] 接口，则在读取数据之后，会调用其接口函数。
func (q *Queries) Object(v any, id string) {
	for k, err := range query.Parse(q.queries, v) {
		q.v.Add(k, localeutil.Phrase(err.Error()))
	}

	if vv, ok := v.(CTXSanitizer); ok {
		if va := vv.CTXSanitize(q.ctx); va != nil {
			va.v.Visit(func(name string, reason localeutil.LocaleStringer) bool {
				q.v.Add(name, reason)
				return true
			})
		}
	}
}

// QueryObject 将查询参数解析到一个对象中
func (ctx *Context) QueryObject(v any, id string) Responser {
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
	if !header.CharsetIsNop(ctx.inputCharset) {
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
// 如果 v 实现了 [CTXSanitizer] 接口，则在读取数据之后，会调用其接口函数。
// 如果验证失败，会输出以 id 作为错误代码的 Response 对象。
func (ctx *Context) Read(v any, id string) Responser {
	if err := ctx.Unmarshal(v); err != nil {
		return ctx.Error(http.StatusUnprocessableEntity, err)
	}

	if vv, ok := v.(CTXSanitizer); ok {
		if va := vv.CTXSanitize(ctx); va != nil {
			return va.Problem(id)
		}
	}

	return nil
}
