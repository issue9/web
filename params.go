// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"strconv"

	"github.com/issue9/mux"
)

var emptyParams = mux.Params(map[string]string{})

// Params 用于处理路径中包含的参数。用法类似于 flag
//  p := ctx.Params()
//  aid := p.Int64("aid")
//  bid := p.Int64("bid")
//  if !p.OK(4000001){
//      return
//  }
type Params struct {
	ctx    *Context
	params mux.Params
	errors map[string]string
}

// Params 声明一个新的 Params 实例
func (ctx *Context) Params() *Params {
	params, err := mux.GetParams(ctx.Request())
	if err != nil {
		ctx.Error(err)
	}
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

// Int 获取参数 key 所代表的值
func (p *Params) Int(key string) int {
	return int(p.Int64(key))
}

// MustInt 获取参数 key 所代表的值。
// 若不存在或是转换出错，则返回 def 作为其默认值。
func (p *Params) MustInt(key string, def int) int {
	return int(p.MustInt64(key, int64(def)))
}

// Int64 获取参数 key 所代表的值
func (p *Params) Int64(key string) int64 {
	ret, err := p.params.Int(key)
	if err != nil {
		p.errors[key] = err.Error()
	}

	return ret
}

// MustInt64 获取参数 key 所代表的值。
// 若不存在或是转换出错，则返回 def 作为其默认值。
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

// String 获取参数 key 所代表的值
func (p *Params) String(key string) string {
	ret, err := p.params.String(key)
	if err != nil {
		p.errors[key] = err.Error()
	}

	return ret
}

// MustString 获取参数 key 所代表的值。
// 若不存在或是转换出错，则返回 def 作为其默认值。
func (p *Params) MustString(key, def string) string {
	ret, found := p.params[key]

	// 不存在，仅返回默认值，不算错误
	if !found {
		return def
	}
	return ret
}

// Bool 获取参数 key 所代表的值
func (p *Params) Bool(key string) bool {
	ret, err := p.params.Bool(key)
	if err != nil {
		p.errors[key] = err.Error()
	}

	return ret
}

// MustBool 获取参数 key 所代表的值。
// 若不存在或是转换出错，则返回 def 作为其默认值。
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

// Float64 获取参数 key 所代表的值
func (p *Params) Float64(key string) float64 {
	ret, err := p.params.Float(key)
	if err != nil {
		p.errors[key] = err.Error()
	}

	return ret
}

// MustFloat64 获取参数 key 所代表的值。
// 若不存在或是转换出错，则返回 def 作为其默认值。
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

// OK 是否一切正常，若出错，则自动向 w 输出 *Result 错误信息，并返回 false
func (p *Params) OK(code int) bool {
	if len(p.errors) > 0 {
		NewResult(code, p.errors).Render(p.ctx)
		return false
	}
	return true
}
