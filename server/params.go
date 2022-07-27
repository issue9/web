// SPDX-License-Identifier: MIT

package server

import (
	"errors"

	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v7/types"

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
		v:   ctx.Server().Problems().NewValidation(),
	}
}

// ID 获取参数 key 所代表的值并转换成 int64
//
// 值必须大于 0，否则会输出错误信息，并返回零值。
func (p *Params) ID(key string) int64 {
	id := p.Int64(key)
	if id <= 0 {
		p.v.Add(key, tGreatThanZero)
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
	id := p.MustInt64(key, def)
	if id <= 0 {
		p.v.Add(key, tGreatThanZero)
		return def
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

// MustInt64 获取参数 key 所代表的值并转换成 int64
//
// 若不存在或是转换出错，则返回 def 作为其默认值。
// 仅在类型转换出错时，才会向 errors 写入错误信息。
func (p *Params) MustInt64(key string, def int64) int64 {
	id, err := p.ctx.route.Params().Int(key)
	if errors.Is(err, types.ErrParamNotExists) { // 不存在，仅返回默认值，不算错误
		return def
	} else if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
		return def
	}
	return id
}

// String 获取参数 key 所代表的值并转换成 string
func (p *Params) String(key string) string {
	ret, err := p.ctx.route.Params().String(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
	}

	return ret
}

// MustString 获取参数 key 所代表的值并转换成 string
//
// 若不存在或是转换出错，则返回 def 作为其默认值。
func (p *Params) MustString(key, def string) string {
	ret, err := p.ctx.route.Params().String(key)
	if errors.Is(err, types.ErrParamNotExists) {
		return def
	} else if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
		return def
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

// MustBool 获取参数 key 所代表的值并转换成 bool
//
// 若不存在或是转换出错，则返回 def 作为其默认值。
// 仅在类型转换出错时，才会向 errors 写入错误信息。
//
// 最终会调用 strconv.ParseBool 进行转换，
// 也只有该方法中允许的字符串会被正确转换。
func (p *Params) MustBool(key string, def bool) bool {
	b, err := p.ctx.route.Params().Bool(key)
	if errors.Is(err, types.ErrParamNotExists) { // 不存在，仅返回默认值，不算错误
		return def
	} else if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
		return def
	}
	return b
}

// Float64 获取参数 key 所代表的值并转换成 float64
func (p *Params) Float64(key string) float64 {
	ret, err := p.ctx.route.Params().Float(key)
	if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
	}

	return ret
}

// MustFloat64 获取参数 key 所代表的值并转换成 float64
//
// 若不存在或是转换出错，则返回 def 作为其默认值。
// 仅在类型转换出错时，才会向 errors 写入错误信息。
func (p *Params) MustFloat64(key string, def float64) float64 {
	f, err := p.ctx.route.Params().Float(key)
	if errors.Is(err, types.ErrParamNotExists) { // 不存在，仅返回默认值，不算错误
		return def
	} else if err != nil {
		p.v.Add(key, localeutil.Phrase(err.Error()))
		return def
	}
	return f
}

func (p *Params) Problem(id string) response.Responser {
	return p.v.Problem(id, p.ctx.LocalePrinter())
}

// ParamID 获取地址参数中表示 key 的值并并转换成大于 0 的 int64
//
// 相对于 Context.ParamInt64()，该值必须大于 0。
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
