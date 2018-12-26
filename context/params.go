// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"errors"
	"strconv"

	"github.com/issue9/mux"
	"github.com/issue9/mux/params"
)

var emptyParams = params.Params(map[string]string{})

// Params 用于处理路径中包含的参数。
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
	errors map[string]string
}

// Params 声明一个新的 Params 实例
func (ctx *Context) Params() *Params {
	params := mux.Params(ctx.Request)
	if params == nil {
		params = emptyParams
	}

	return &Params{
		ctx:    ctx,
		params: params,
		errors: make(map[string]string, len(params)),
	}
}

// ID 获取参数 key 所代表的值，并转换成 int 且值必须大于 0。
func (p *Params) ID(key string) int64 {
	id := p.Int64(key)
	if id <= 0 {
		p.errors[key] = "必须大于 0"
	}

	return id
}

// MustID 获取参数 key 所代表的值，转换成 int64 且必须大于 0。
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
		p.errors[key] = err.Error()
		return def
	}

	if ret <= 0 {
		p.errors[key] = "必须大于 0"
		return def
	}

	return ret
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

// HasErrors 是否有错误内容存在
func (p *Params) HasErrors() bool {
	return len(p.errors) > 0
}

// Errors 返回所有的错误信息
func (p *Params) Errors() map[string]string {
	return p.errors
}

// Result 转换成 Result 对象
//
// code 是作为 Result.Code 从错误消息中查找，如果不存在，则 panic。
// Params.errors 将会作为 Result.Detail 的内容。
func (p *Params) Result(code int) *Result {
	return p.ctx.NewResult(code).SetDetail(p.Errors())
}

// ParamID 获取地址参数中表示 ID 的值。相对于 ParamInt64，该值必须大于 0。
//
// NOTE: 若需要获取多个参数，可以使用 Context.Params 获取会更方便。
func (ctx *Context) ParamID(key string) (int64, error) {
	id, err := ctx.ParamInt64(key)

	if err != nil {
		return 0, err
	}

	if id <= 0 {
		return 0, errors.New("必须大于 0")
	}

	return id, nil
}

// ParamInt64 取地址参数中的 int64 值。
//
// NOTE: 若需要获取多个参数，可以使用 Context.Params 获取会更方便。
func (ctx *Context) ParamInt64(key string) (int64, error) {
	return ctx.Params().params.Int(key)
}

// ParamString 取地址参数中的 string 值。
//
// NOTE: 若需要获取多个参数，可以使用 Context.Params 获取会更方便。
func (ctx *Context) ParamString(key string) (string, error) {
	return ctx.Params().params.String(key)
}
