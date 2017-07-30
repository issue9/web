// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"strconv"

	"github.com/issue9/mux"
	"github.com/issue9/mux/params"
)

var emptyParams = params.Params(map[string]string{})

// Params 用于处理路径中包含的参数。
//  p := ctx.Params()
//  aid := p.Int64("aid")
//  bid := p.Int64("bid")
//  if !p.OK(4000001){
//      return
//  }
type Params struct {
	ctx    *Context
	params params.Params
	errors map[string]string
}

// Params 声明一个新的 Params 实例
func (ctx *Context) Params() *Params {
	params := mux.Params(ctx.Request())
	if params == nil {
		ctx.Error("mux.GetParams() 中获取的值为一个空的 map")
		params = emptyParams
	}

	return &Params{
		ctx:    ctx,
		params: params,
		errors: make(map[string]string, len(params)),
	}
}

// Int 获取参数 key 所代表的值，并转换成 int。
func (p *Params) Int(key string) int {
	return int(p.Int64(key))
}

// MustInt 获取参数 key 所代表的值，并转换成 int。
// 若不存在或是转换出错，则返回 def 作为其默认值。
// 仅在类型转换出错时，才会向 errors 写入错误信息。
func (p *Params) MustInt(key string, def int) int {
	return int(p.MustInt64(key, int64(def)))
}

// Int64 获取参数 key 所代表的值，并转换成 int64。
func (p *Params) Int64(key string) int64 {
	ret, err := p.params.Int(key)
	if err != nil {
		p.errors[key] = err.Error()
	}

	return ret
}

// MustInt64 获取参数 key 所代表的值，并转换成 int64。
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
		p.errors[key] = err.Error()
		return def
	}

	return ret
}

// String 获取参数 key 所代表的值，并转换成 string。
func (p *Params) String(key string) string {
	ret, err := p.params.String(key)
	if err != nil {
		p.errors[key] = err.Error()
	}

	return ret
}

// MustString 获取参数 key 所代表的值，并转换成 string。
// 若不存在或是转换出错，则返回 def 作为其默认值。
func (p *Params) MustString(key, def string) string {
	ret, found := p.params[key]

	// 不存在，仅返回默认值，不算错误
	if !found {
		return def
	}
	return ret
}

// Bool 获取参数 key 所代表的值，并转换成 bool。
//
// 最终会调用 strconv.ParseBool 进行转换，
// 也只有该方法中允许的字符串会被正确转换。
func (p *Params) Bool(key string) bool {
	ret, err := p.params.Bool(key)
	if err != nil {
		p.errors[key] = err.Error()
	}

	return ret
}

// MustBool 获取参数 key 所代表的值，并转换成 bool。
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
		p.errors[key] = err.Error()
		return def
	}

	return ret
}

// Float64 获取参数 key 所代表的值，并转换成 float64。
func (p *Params) Float64(key string) float64 {
	ret, err := p.params.Float(key)
	if err != nil {
		p.errors[key] = err.Error()
	}

	return ret
}

// MustFloat64 获取参数 key 所代表的值，并转换成 float64。
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
		p.errors[key] = err.Error()
		return def
	}

	return ret
}

// Result 获取解析的结果，若存在错误则返回相应的 *Result 实例
func (p *Params) Result(code int) *Result {
	if len(p.errors) == 0 {
		return nil
	}

	return NewResult(code, p.errors)
}

// OK 是否一切正常，若出错，则自动向客户端输出 *Result 错误信息，并返回 false
func (p *Params) OK(code int) bool {
	if len(p.errors) > 0 {
		NewResult(code, p.errors).Render(p.ctx)
		return false
	}
	return true
}

// ParamID 获取地址参数中表示 ID 的值。相对于 int64，但该值必须大于 0。
// 当出错时，第二个参数返回 false。
//
// NOTE: 若需要获取其它类型的数据，可以使用 Context.Params 来获取。
// Context.ParamInt64 和 Context.ParamID 仅作为一个简便的操作存在。
func (ctx *Context) ParamID(key string, code int) (int64, bool) {
	p := ctx.Params()
	id := p.Int64(key)

	if !p.OK(code) {
		return 0, false
	}

	if id <= 0 {
		NewResult(code, map[string]string{key: "必须大于零"}).Render(ctx)
		return 0, false
	}

	return id, true
}

// ParamInt64 取地址参数中的 int64 值。
// 当出错时，第二个参数返回 false。
//
// NOTE: 若需要获取其它类型的数据，可以使用 Context.Params 来获取。
// Context.ParamInt64 和 Context.ParamID 仅作为一个简便的操作存在。
func (ctx *Context) ParamInt64(key string, code int) (int64, bool) {
	p := ctx.Params()
	id := p.Int64(key)

	if !p.OK(code) {
		return id, false
	}

	return id, true
}
