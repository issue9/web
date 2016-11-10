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

// Param
type Param struct {
	abortOnError bool
	errors       map[string]string
	values       map[string]value
	params       map[string]string // 从 context 中获取的参数列表
}

func NewParam(r *http.Request, abortOnError bool) (*Param, error) {
	var params map[string]string
	m, found := context.Get(r).Get("params")
	if !found {
		return nil, errors.New("context 中的不存在 params 参数")
	}

	var ok bool
	params, ok = m.(map[string]string)
	if !ok {
		return nil, errors.New("无法将 context 中的 params 参数转换成 map[string]string 类型")
	}

	return &Param{
		abortOnError: abortOnError,
		errors:       map[string]string{},
		values:       make(map[string]value, len(params)),
		params:       params,
	}, nil
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

func (p *Param) ID(key string) *int64 {
	i := new(int64)
	p.IDVar(i, key)
	return i
}

func (p *Param) IDVar(i *int64, key string) {
	p.values[key] = (*int64Value)(i)
}

func (p *Param) Int64(key string) *int64 {
	i := new(int64)
	p.Int64Var(i, key)
	return i
}

func (p *Param) Int64Var(i *int64, key string) {
	p.values[key] = (*int64Value)(i)
}

func (p *Param) String(key string) *string {
	i := new(string)
	p.StringVar(i, key)
	return i
}

func (p *Param) StringVar(i *string, key string) {
	p.values[key] = (*stringValue)(i)
}

func (p *Param) Bool(key string) *bool {
	i := new(bool)
	p.BoolVar(i, key)
	return i
}

func (p *Param) BoolVar(i *bool, key string) {
	p.values[key] = (*boolValue)(i)
}

func (p *Param) Int(key string) *int {
	i := new(int)
	p.IntVar(i, key)
	return i
}

func (p *Param) IntVar(i *int, key string) {
	p.values[key] = (*intValue)(i)
}

func (p *Param) Parse() map[string]string {
	for k, v := range p.values {
		ok := p.parseOne(k, v)
		if !ok && p.abortOnError {
			break
		}
	}

	return p.errors
}
