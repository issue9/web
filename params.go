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

func (p *Params) parseOne(key string, val value) {
	v, found := p.params[key]
	if !found {
		p.errors[key] = "参数名不存在"
		return
	}

	if err := val.set(v); err != nil {
		p.errors[key] = err.Error()
	}
	return
}

// Int 获取参数 key 所代表的值
func (p *Params) Int(key string) int {
	i := new(int)
	p.intVar(i, key)
	return *i
}

// IntVar 将参数 key 所代表的值保存到指针 i 中
func (p *Params) IntVar(i *int, key string) *Params {
	p.intVar(i, key)
	return p
}

func (p *Params) intVar(i *int, key string) {
	val := (*intValue)(i)
	p.parseOne(key, val)
}

// Int64 获取参数 key 所代表的值
func (p *Params) Int64(key string) int64 {
	i := new(int64)
	p.int64Var(i, key)
	return *i
}

// Int64Var 将参数 key 所代表的值保存到指针 i 中
func (p *Params) Int64Var(i *int64, key string) *Params {
	p.int64Var(i, key)
	return p
}

func (p *Params) int64Var(i *int64, key string) {
	val := (*int64Value)(i)
	p.parseOne(key, val)
}

// String 获取参数 key 所代表的值
func (p *Params) String(key string) string {
	i := new(string)
	p.stringVar(i, key)
	return *i
}

// StringVar 将参数 key 所代表的值保存到指针 i 中
func (p *Params) StringVar(i *string, key string) *Params {
	p.stringVar(i, key)
	return p
}

func (p *Params) stringVar(i *string, key string) {
	val := (*stringValue)(i)
	p.parseOne(key, val)
}

// Bool 获取参数 key 所代表的值
func (p *Params) Bool(key string) bool {
	i := new(bool)
	p.boolVar(i, key)
	return *i
}

// BoolVar 将参数 key 所代表的值保存到指针 i 中
func (p *Params) BoolVar(i *bool, key string) *Params {
	p.boolVar(i, key)
	return p
}

func (p *Params) boolVar(i *bool, key string) {
	val := (*boolValue)(i)
	p.parseOne(key, val)
}

// Float64 获取参数 key 所代表的值
func (p *Params) Float64(key string) float64 {
	i := new(float64)
	p.float64Var(i, key)
	return *i
}

// Float64Var 将参数 key 所代表的值保存到指针 i 中
func (p *Params) Float64Var(i *float64, key string) *Params {
	p.float64Var(i, key)
	return p
}

func (p *Params) float64Var(i *float64, key string) {
	val := (*float64Value)(i)
	p.parseOne(key, val)
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
