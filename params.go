// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import "github.com/issue9/mux"

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

func (p *Params) MustInt(key string, def int) int {
	return int(p.params.MustInt(key, int64(def)))
}

// Int64 获取参数 key 所代表的值
func (p *Params) Int64(key string) int64 {
	ret, err := p.params.Int(key)
	if err != nil {
		p.errors[key] = err.Error()
	}

	return ret
}

func (p *Params) MustInt64(key string, def int64) int64 {
	return p.params.MustInt(key, def)
}

// String 获取参数 key 所代表的值
func (p *Params) String(key string) string {
	ret, err := p.params.String(key)
	if err != nil {
		p.errors[key] = err.Error()
	}

	return ret
}

func (p *Params) MustString(key, def string) string {
	return p.params.MustString(key, def)
}

// Bool 获取参数 key 所代表的值
func (p *Params) Bool(key string) bool {
	ret, err := p.params.Bool(key)
	if err != nil {
		p.errors[key] = err.Error()
	}

	return ret
}

// MustBool
func (p *Params) MustBool(key string, def bool) bool {
	return p.params.MustBool(key, def)
}

// Float64 获取参数 key 所代表的值
func (p *Params) Float64(key string) float64 {
	ret, err := p.params.Float(key)
	if err != nil {
		p.errors[key] = err.Error()
	}

	return ret
}

func (p *Params) MustFloat64(key string, def float64) float64 {
	return p.params.MustFloat(key, def)
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
		p.Result(code).Render(p.ctx)
		return false
	}
	return true
}
