// SPDX-License-Identifier: MIT

package server

import (
	"strconv"

	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/params"
)

var emptyParams = params.Params(map[string]string{})

// Params 用于处理路径中包含的参数
//  p := ctx.Params()
//  aid := p.Int64("aid")
//  bid := p.Int64("bid")
//  if p.HasErrors() {
//      // do something
//      return
//  }
type Params struct {
	ctx    *Context
	params params.Params
	errors ResultFields
}

// Params 声明一个新的 Params 实例
func (ctx *Context) Params() *Params {
	ps := mux.Params(ctx.Request)
	if ps == nil {
		ps = emptyParams
	}

	return &Params{
		ctx:    ctx,
		params: ps,
		errors: make(ResultFields, len(ps)),
	}
}

// ID 获取参数 key 所代表的值并转换成 int64
//
// 值必须大于 0，否则会输出错误信息，并返回零值。
func (p *Params) ID(key string) int64 {
	id := p.Int64(key)
	if id <= 0 {
		p.errors.Add(key, p.ctx.LocalePrinter.Sprintf("should great than 0"))
	}

	return id
}

// MustID 获取参数 key 所代表的值并转换成 int64
//
// 值必须大于 0，否则会输出错误信息，并返回零值。
//
// 若不存在或是转换出错，则返回 def 作为其默认值。
// 仅在类型转换出错或是小于零时，才会向 errors 写入错误信息。
func (p *Params) MustID(key string, def int64) int64 {
	str, found := p.params[key]

	// 不存在，仅返回默认值，不算错误
	if !found {
		return def
	}

	ret, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		p.errors.Add(key, err.Error())
		return def
	}

	if ret <= 0 {
		p.errors.Add(key, p.ctx.LocalePrinter.Sprintf("should great than 0"))
		return def
	}

	return ret
}

// Int64 获取参数 key 所代表的值，并转换成 int64
func (p *Params) Int64(key string) int64 {
	ret, err := p.params.Int(key)
	if err != nil {
		p.errors.Add(key, err.Error())
	}

	return ret
}

// MustInt64 获取参数 key 所代表的值并转换成 int64
//
// 若不存在或是转换出错，则返回 def 作为其默认值。
// 仅在类型转换出错时，才会向 errors 写入错误信息。
func (p *Params) MustInt64(key string, def int64) int64 {
	str, found := p.params[key]

	// 不存在，仅返回默认值，不算错误
	if !found {
		return def
	}

	ret, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		p.errors.Add(key, err.Error())
		return def
	}

	return ret
}

// String 获取参数 key 所代表的值并转换成 string
func (p *Params) String(key string) string {
	ret, err := p.params.String(key)
	if err != nil {
		p.errors.Add(key, err.Error())
	}

	return ret
}

// MustString 获取参数 key 所代表的值并转换成 string
//
// 若不存在或是转换出错，则返回 def 作为其默认值。
func (p *Params) MustString(key, def string) string {
	ret, found := p.params[key]

	// 不存在，仅返回默认值，不算错误
	if !found {
		return def
	}
	return ret
}

// Bool 获取参数 key 所代表的值并转换成 bool
//
// 最终会调用 strconv.ParseBool 进行转换，
// 也只有该方法中允许的字符串会被正确转换。
func (p *Params) Bool(key string) bool {
	ret, err := p.params.Bool(key)
	if err != nil {
		p.errors.Add(key, err.Error())
	}

	return ret
}

// MustBool 获取参数 key 所代表的值并转换成 bool
//
// 若不存在或是转换出错，则返回 def 作为其默认值。
// 仅在类型转换出错时，才会向 errors 写入错误信息。
//
// 最终会调用 strconv.ParseBool 进行转换，
// 也只有该方法中允许的字符串会被正确转换。
func (p *Params) MustBool(key string, def bool) bool {
	str, found := p.params[key]

	// 不存在，仅返回默认值，不算错误
	if !found {
		return def
	}

	ret, err := strconv.ParseBool(str)
	if err != nil {
		p.errors.Add(key, err.Error())
		return def
	}

	return ret
}

// Float64 获取参数 key 所代表的值并转换成 float64
func (p *Params) Float64(key string) float64 {
	ret, err := p.params.Float(key)
	if err != nil {
		p.errors.Add(key, err.Error())
	}

	return ret
}

// MustFloat64 获取参数 key 所代表的值并转换成 float64
//
// 若不存在或是转换出错，则返回 def 作为其默认值。
// 仅在类型转换出错时，才会向 errors 写入错误信息。
func (p *Params) MustFloat64(key string, def float64) float64 {
	str, found := p.params[key]

	// 不存在，仅返回默认值，不算错误
	if !found {
		return def
	}

	ret, err := strconv.ParseFloat(str, 64)
	if err != nil {
		p.errors.Add(key, err.Error())
		return def
	}

	return ret
}

// HasErrors 是否有错误内容存在
func (p *Params) HasErrors() bool { return len(p.errors) > 0 }

// Errors 返回所有的错误信息
func (p *Params) Errors() ResultFields { return p.errors }

// Result 转换成 Result 对象
func (p *Params) Result(code int) Responser {
	if p.HasErrors() {
		return p.ctx.Result(code, p.Errors())
	}
	return nil
}

// ParamID 获取地址参数中表示 key 的值并并转换成大于 0 的 int64
//
// 相对于 Context.ParamInt64()，该值必须大于 0。
//
// NOTE: 若需要获取多个参数，使用 Context.Params 会更方便。
func (ctx *Context) ParamID(key string, code int) (int64, Responser) {
	id, resp := ctx.ParamInt64(key, code)
	if resp != nil {
		return 0, resp
	}

	if id <= 0 {
		return 0, ctx.Result(code, ResultFields{key: []string{"必须大于 0"}})
	}
	return id, nil
}

// ParamInt64 取地址参数中的 key 表示的值 int64 类型值
//
// NOTE: 若需要获取多个参数，可以使用 Context.Params 获取会更方便。
func (ctx *Context) ParamInt64(key string, code int) (int64, Responser) {
	v, err := ctx.Params().params.Int(key)
	if err != nil {
		return 0, ctx.Result(code, ResultFields{key: []string{err.Error()}})
	}
	return v, nil
}

// ParamString 取地址参数中的 key 表示的 string 类型值
//
// NOTE: 若需要获取多个参数，可以使用 Context.Params 获取会更方便。
func (ctx *Context) ParamString(key string, code int) (string, Responser) {
	v, err := ctx.Params().params.String(key)
	if err != nil {
		return "", ctx.Result(code, ResultFields{key: []string{err.Error()}})
	}
	return v, nil
}
