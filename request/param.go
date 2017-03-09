// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package request

import (
	"errors"
	"net/http"

	"github.com/issue9/mux"
	"github.com/issue9/web/context"
	"github.com/issue9/web/result"
)

const paramNotExists = "参数不存在"

var emptyParams = map[string]string{}

// Param 用于处理路径中包含的参数。用法类似于 flag
//  p := NewParam(false)
//  aid := p.Int64("aid")
//  bid := p.Int64("bid")
//  if !p.OK(w){
//      return
//  }
type Param struct {
	errors map[string]string
	params map[string]string // 从 context 中获取的参数列表
}

// NewParam 声明一个新的 Param 实例
//
// NOTE:当出错时，会返回一个空的 Param 实例，而不是 nil
func NewParam(r *http.Request) (*Param, error) {
	params := r.Context().Value(mux.ContextKeyParams)
	if params == nil {
		return newParam(emptyParams), errors.New("r.Context() 中未包含有关参数的信息")
	}

	m, ok := params.(mux.Params)
	if !ok {
		return newParam(emptyParams), errors.New("从 r.Context() 中获取的值无法正确转换到 mux.Params")
	}

	if m == nil {
		return newParam(emptyParams), errors.New("从 r.Context() 中获取的值为一个空的 map")
	}

	return newParam(m), nil
}

func newParam(params map[string]string) *Param {
	return &Param{
		errors: make(map[string]string, len(params)),
		params: params,
	}
}

func (p *Param) parseOne(key string, val value) {
	v, found := p.params[key]
	if !found {
		// NOTE: 能导航到此，又找不到参数，说明代码逻辑有问题。
		p.errors[key] = paramNotExists
		return
	}

	if err := val.set(v); err != nil {
		p.errors[key] = err.Error()
	}
	return
}

// Int 获取参数 key 所代表的值
func (p *Param) Int(key string) int {
	i := new(int)
	p.intVar(i, key)
	return *i
}

// IntVar 将参数 key 所代表的值保存到指针 i 中
func (p *Param) IntVar(i *int, key string) *Param {
	p.intVar(i, key)
	return p
}

func (p *Param) intVar(i *int, key string) {
	val := (*intValue)(i)
	p.parseOne(key, val)
}

// Int64 获取参数 key 所代表的值
func (p *Param) Int64(key string) int64 {
	i := new(int64)
	p.int64Var(i, key)
	return *i
}

// Int64Var 将参数 key 所代表的值保存到指针 i 中
func (p *Param) Int64Var(i *int64, key string) *Param {
	p.int64Var(i, key)
	return p
}

func (p *Param) int64Var(i *int64, key string) {
	val := (*int64Value)(i)
	p.parseOne(key, val)
}

// String 获取参数 key 所代表的值
func (p *Param) String(key string) string {
	i := new(string)
	p.stringVar(i, key)
	return *i
}

// StringVar 将参数 key 所代表的值保存到指针 i 中
func (p *Param) StringVar(i *string, key string) *Param {
	p.stringVar(i, key)
	return p
}

func (p *Param) stringVar(i *string, key string) {
	val := (*stringValue)(i)
	p.parseOne(key, val)
}

// Bool 获取参数 key 所代表的值
func (p *Param) Bool(key string) bool {
	i := new(bool)
	p.boolVar(i, key)
	return *i
}

// BoolVar 将参数 key 所代表的值保存到指针 i 中
func (p *Param) BoolVar(i *bool, key string) *Param {
	p.boolVar(i, key)
	return p
}

func (p *Param) boolVar(i *bool, key string) {
	val := (*boolValue)(i)
	p.parseOne(key, val)
}

// Float64 获取参数 key 所代表的值
func (p *Param) Float64(key string) float64 {
	i := new(float64)
	p.float64Var(i, key)
	return *i
}

// Float64Var 将参数 key 所代表的值保存到指针 i 中
func (p *Param) Float64Var(i *float64, key string) *Param {
	p.float64Var(i, key)
	return p
}

func (p *Param) float64Var(i *float64, key string) {
	val := (*float64Value)(i)
	p.parseOne(key, val)
}

// Result 获取解析的结果，若存在错误则返回相应的 *result.Result 实例
func (p *Param) Result(code int) *result.Result {
	if len(p.errors) == 0 {
		return result.New(code)
	}

	return result.NewWithDetail(code, p.errors)
}

// OK 是否一切正常，若出错，则自动向 w 输出错误信息，并返回 false
func (p *Param) OK(ctx context.Context, code int) bool {
	if len(p.errors) > 0 {
		p.Result(code).Render(ctx)
		return false
	}
	return true
}
