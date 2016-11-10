// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package request

import (
	"errors"
	"net/http"

	"github.com/issue9/context"
)

const paramNotExists = "参数不存在"

// Param 用于处理路径中包含的参数。用法类似于 flag
//  p,_ := NewParam(r, false)
//  aid := p.Int64("aid")
//  bid := p.Int64("bid")
//  if msg := p.Parse();len(msg)>0{
//      rslt := web.NewResultWithDetail(400, msg)
//      rslt.RenderJSON(w,r,nil)
//      return
//  }
type Param struct {
	abortOnError bool
	errors       map[string]string
	values       map[string]value
	params       map[string]string // 从 context 中获取的参数列表
}

// NewParam 声明一个新的 Param 实例
//
// NOTE:当出错时，会返回一个空的 Param 实例，而不是 nil
func NewParam(r *http.Request, abortOnError bool) (*Param, error) {
	var params map[string]string
	var err error
	m, found := context.Get(r).Get("params")
	if !found {
		return newParam(params, abortOnError), errors.New("context 中的不存在 params 参数")
	}

	var ok bool
	params, ok = m.(map[string]string)
	if !ok {
		return newParam(params, abortOnError), errors.New("无法将 context 中的 params 参数转换成 map[string]string 类型")
	}

	return newParam(params, abortOnError), err
}

func newParam(params map[string]string, abortOnError bool) *Param {
	return &Param{
		abortOnError: abortOnError,
		errors:       make(map[string]string, len(params)),
		values:       make(map[string]value, len(params)),
		params:       params,
	}
}

func (p *Param) parseOne(key string, val value) (ok bool) {
	v, found := p.params[key]
	if !found {
		p.errors[key] = paramNotExists
		return false
	}

	if err := val.set(v); err != nil {
		p.errors[key] = err.Error()
		return false
	}
	return true
}

// ID 获取参数 key 所代表的值，其值必须大于 0
func (p *Param) ID(key string) *int64 {
	i := new(int64)
	p.IDVar(i, key)
	return i
}

// IDVar 将参数 key 所代表的值保存到指针 i 中，其值必须大于 0
func (p *Param) IDVar(i *int64, key string) {
	p.values[key] = (*int64Value)(i)
}

// Int 获取参数 key 所代表的值
func (p *Param) Int(key string) *int {
	i := new(int)
	p.IntVar(i, key)
	return i
}

// IntVar 将参数 key 所代表的值保存到指针 i 中
func (p *Param) IntVar(i *int, key string) {
	p.values[key] = (*intValue)(i)
}

// Int64 获取参数 key 所代表的值
func (p *Param) Int64(key string) *int64 {
	i := new(int64)
	p.Int64Var(i, key)
	return i
}

// Int64Var 将参数 key 所代表的值保存到指针 i 中
func (p *Param) Int64Var(i *int64, key string) {
	p.values[key] = (*int64Value)(i)
}

// String 获取参数 key 所代表的值
func (p *Param) String(key string) *string {
	i := new(string)
	p.StringVar(i, key)
	return i
}

// StringVar 将参数 key 所代表的值保存到指针 i 中
func (p *Param) StringVar(i *string, key string) {
	p.values[key] = (*stringValue)(i)
}

// Bool 获取参数 key 所代表的值
func (p *Param) Bool(key string) *bool {
	i := new(bool)
	p.BoolVar(i, key)
	return i
}

// BoolVar 将参数 key 所代表的值保存到指针 i 中
func (p *Param) BoolVar(i *bool, key string) {
	p.values[key] = (*boolValue)(i)
}

// Float64 获取参数 key 所代表的值
func (p *Param) Float64(key string) *float64 {
	i := new(float64)
	p.Float64Var(i, key)
	return i
}

// Float64Var 将参数 key 所代表的值保存到指针 i 中
func (p *Param) Float64Var(i *float64, key string) {
	p.values[key] = (*float64Value)(i)
}

// Parse 开始解析数据，若存在错误则返回每个参数对应的错误信息
func (p *Param) Parse() map[string]string {
	for k, v := range p.values {
		ok := p.parseOne(k, v)
		if !ok && p.abortOnError {
			break
		}
	}

	return p.errors
}
