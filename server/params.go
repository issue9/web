// SPDX-License-Identifier: MIT

package server

import (
	"github.com/issue9/localeutil"

	"github.com/issue9/web/problem"
	"github.com/issue9/web/server/response"
)

var tGreatThanZero = localeutil.Phrase("should great than 0")

// Params 用于处理路径中包含的参数
//  p := ctx.Params()
//  aid := p.Int64("aid")
//  bid := p.Int64("bid")
type Params struct {
	ctx *Context
	v   *problem.Validation
}

// Params 声明一个新的 Params 实例
func (ctx *Context) Params() *Params {
	return &Params{
		ctx: ctx,
		v:   ctx.NewValidation(),
	}
}

// ID 获取参数 key 所代表的值并转换成 int64
//
// 值必须大于 0，否则会输出错误信息，并返回零值。
func (p *Params) ID(key string) int64 {
	id, err := p.ctx.route.Params().Int(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
	} else if id <= 0 {
		p.v.Add(key, tGreatThanZero)
	}
	return id
}

// Int64 获取参数 key 所代表的值，并转换成 int64
func (p *Params) Int64(key string) int64 {
	ret, err := p.ctx.route.Params().Int(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
	}
	return ret
}

// String 获取参数 key 所代表的值并转换成 string
func (p *Params) String(key string) string {
	ret, err := p.ctx.route.Params().String(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
	}
	return ret
}

// Bool 获取参数 key 所代表的值并转换成 bool
//
// 最终会调用 strconv.ParseBool 进行转换，
// 也只有该方法中允许的字符串会被正确转换。
func (p *Params) Bool(key string) bool {
	ret, err := p.ctx.route.Params().Bool(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
	}
	return ret
}

// Float64 获取参数 key 所代表的值并转换成 float64
func (p *Params) Float64(key string) float64 {
	ret, err := p.ctx.route.Params().Float(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
	}
	return ret
}

func (p *Params) Problem(id string) response.Responser {
	return p.v.Problem(p.ctx.Server().Problems(), id, p.ctx.LocalePrinter())
}

// ParamID 获取地址参数中表示 key 的值并并转换成大于 0 的 int64
//
// NOTE: 若需要获取多个参数，使用 Context.Params 会更方便。
func (ctx *Context) ParamID(key, id string) (int64, response.Responser) {
	p := ctx.Params()
	num := p.ID(key)
	if pp := p.Problem(id); pp != nil {
		return 0, pp
	}
	return num, nil
}

// ParamInt64 取地址参数中的 key 表示的值 int64 类型值
//
// NOTE: 若需要获取多个参数，可以使用 Context.Params 获取会更方便。
func (ctx *Context) ParamInt64(key, id string) (int64, response.Responser) {
	p := ctx.Params()
	num := p.Int64(key)
	if pp := p.Problem(id); pp != nil {
		return 0, pp
	}
	return num, nil
}

// ParamString 取地址参数中的 key 表示的 string 类型值
//
// NOTE: 若需要获取多个参数，可以使用 Context.Params 获取会更方便。
func (ctx *Context) ParamString(key, id string) (string, response.Responser) {
	p := ctx.Params()
	s := p.String(key)
	if pp := p.Problem(id); pp != nil {
		return "", pp
	}
	return s, nil
}
