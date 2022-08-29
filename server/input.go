// SPDX-License-Identifier: MIT

package server

import (
	"io"
	"net/url"
	"strconv"
	"sync"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
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
		v *Validation
	}

	Queries struct {
		v       *Validation
		queries url.Values
	}
)

// Params 声明一个用于获取路径参数的对象
//
// 返回对象的生命周期在 Context 结束时也随之结束。
func (ctx *Context) Params(exitAtError bool) *Params {
	ps := paramPool.Get().(*Params)
	ps.v = ctx.NewValidation(exitAtError)
	ctx.OnExit(func(i int) { paramPool.Put(ps) })
	return ps
}

// ID 返回 key 所表示的值且必须大于 0
func (p *Params) ID(key string) int64 {
	if !p.v.continueNext() {
		return 0
	}

	id, err := p.v.ctx.route.Params().Int(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
		return 0
	} else if id <= 0 {
		p.v.Add(key, tGreatThanZero)
		return 0
	}
	return id
}

// Int64 获取参数 key 所代表的值并转换成 int64 类型
func (p *Params) Int64(key string) int64 {
	if !p.v.continueNext() {
		return 0
	}

	ret, err := p.v.ctx.route.Params().Int(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
		return 0
	}
	return ret
}

// String 获取参数 key 所代表的值并转换成 string
func (p *Params) String(key string) string {
	if !p.v.continueNext() {
		return ""
	}

	ret, err := p.v.ctx.route.Params().String(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
		return ""
	}
	return ret
}

// Bool 获取参数 key 所代表的值并转换成 bool
//
// 由 [strconv.ParseBool] 进行转换。
func (p *Params) Bool(key string) bool {
	if !p.v.continueNext() {
		return false
	}

	ret, err := p.v.ctx.route.Params().Bool(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
	}
	return ret
}

// Float64 获取参数 key 所代表的值并转换成 float64
func (p *Params) Float64(key string) float64 {
	if !p.v.continueNext() {
		return 0
	}

	ret, err := p.v.ctx.route.Params().Float(key)
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
	// 不复用 Params 实例，省略了 Params 和  Validation 两个对象的创建。
	p := ctx.LocalePrinter()
	ret, err := ctx.route.Params().Int(key)
	if err != nil {
		return 0, ctx.Problem(id).AddParam(key, localeutil.Phrase(err.Error()).LocaleString(p))
	} else if ret <= 0 {
		return 0, ctx.Problem(id).AddParam(key, tGreatThanZero.LocaleString(p))
	}
	return ret, nil
}

// ParamInt64 取地址参数中的 key 表示的值并尝试工转换成 int64 类型
//
// NOTE: 若需要获取多个参数，可以使用 [Context.Params] 获取会更方便。
func (ctx *Context) ParamInt64(key, id string) (int64, Responser) {
	// 不复用 Params 实例，省略了 Params 和  Validation 两个对象的创建。
	ret, err := ctx.route.Params().Int(key)
	if err != nil {
		msg := localeutil.Phrase(err.Error()).LocaleString(ctx.LocalePrinter())
		return 0, ctx.Problem(id).AddParam(key, msg)
	}
	return ret, nil
}

// ParamString 取地址参数中的 key 表示的值并尝试工转换成 string 类型
//
// NOTE: 若需要获取多个参数，可以使用 [Context.Params] 获取会更方便。
func (ctx *Context) ParamString(key, id string) (string, Responser) {
	// 不复用 Params 实例，省略了 Params 和  Validation 两个对象的创建。
	ret, err := ctx.route.Params().String(key)
	if err != nil {
		msg := localeutil.Phrase(err.Error()).LocaleString(ctx.LocalePrinter())
		return "", ctx.Problem(id).AddParam(key, msg)
	}
	return ret, nil
}

// Queries 声明一个用于获取查询参数的对象
//
// 返回对象的生命周期在 Context 结束时也随之结束。
func (ctx *Context) Queries(exitAtError bool) (*Queries, error) {
	queries, err := url.ParseQuery(ctx.Request().URL.RawQuery)
	if err != nil {
		return nil, err
	}

	q := queryPool.Get().(*Queries)
	q.v = ctx.NewValidation(exitAtError)
	q.queries = queries
	ctx.OnExit(func(i int) { queryPool.Put(q) })
	return q, nil
}

func (q *Queries) getItem(key string) (val string) { return q.queries.Get(key) }

// Int 从查询参数中获取指定名称的值
//
// 若不存在则返回 def 作为其默认值。若是无法转换，则会保存错误信息并返回 def。
func (q *Queries) Int(key string, def int) int {
	if !q.v.continueNext() {
		return 0
	}

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
	if !q.v.continueNext() {
		return 0
	}

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
	if !q.v.continueNext() {
		return ""
	}

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
	if !q.v.continueNext() {
		return false
	}

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
	if !q.v.continueNext() {
		return 0
	}

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
	query.ParseWithLog(q.queries, v, func(s string, err error) {
		q.v.Add(s, localeutil.Phrase(err.Error()))
	})

	if q.v.continueNext() {
		if s, ok := v.(CTXSanitizer); ok {
			s.CTXSanitize(q.v.ctx, q.v)
		}
	}
}

// QueryObject 将查询参数解析到一个对象中
func (ctx *Context) QueryObject(exitAtError bool, v any, id string) Responser {
	q, err := ctx.Queries(exitAtError)
	if err != nil {
		return ctx.Error(id, logs.LevelError, err)
	}
	q.Object(v, id)

	return q.Problem(id)
}

// Body 获取用户提交的内容
//
// 相对于 Context.Request().Body，此函数可多次读取。不存在 body 时，返回 nil
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
	if ctx.inputMimetype == nil {
		return localeutil.Error("the client did not specify content-type header")
	}
	return ctx.inputMimetype(body, v)
}

// Read 从客户端读取数据并转换成 v 对象
//
// 如果 v 实现了 [CTXSanitizer] 接口，则在读取数据之后，会调用其接口函数。
// 如果验证失败，会输出以 id 作为错误代码的 [Responser] 对象。
func (ctx *Context) Read(exitAtError bool, v any, id string) Responser {
	if err := ctx.Unmarshal(v); err != nil {
		return ctx.Error("422", logs.LevelError, err) // http.StatusUnprocessableEntity
	}

	if vv, ok := v.(CTXSanitizer); ok {
		va := ctx.NewValidation(exitAtError)
		vv.CTXSanitize(ctx, va)
		return va.Problem(id)
	}
	return nil
}
