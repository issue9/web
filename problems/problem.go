// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package problems

import (
	"slices"
	"unique"

	"github.com/issue9/localeutil"
)

type ID = unique.Handle[string]

// Problem 基于 [RFC7807] 和 [RFC9457] 用于描述错误信息的对象
//
// 这是一个提供给客户端的信息，不能将敏感信息放于其中。
//
// [RFC7807]: https://www.rfc-editor.org/info/rfc7807
// [RFC9457]: https://www.rfc-editor.org/info/rfc9457
type Problem[T renderer] struct {
	XMLName struct{} `xml:"urn:ietf:rfc:7807 problem" form:"-" cbor:"-" json:"-" html:"-" yaml:"-"`

	// 错误 ID
	//
	// 如果该值为 about:blank，则表示 HTTP 状态码本身语义已经可以表达此错误。
	Type string `json:"type" xml:"type" form:"type" cbor:"type" yaml:"type" comment:"problem type"`

	// 对错误的简要描述
	Title string `json:"title" xml:"title" form:"title" cbor:"title" yaml:"title" comment:"problem title"`

	// 对错误的详细说明
	Detail string `json:"detail,omitempty" xml:"detail,omitempty" form:"detail,omitempty" cbor:"detail,omitempty" yaml:"detail,omitempty" comment:"problem detail"`

	// 指向错误发生的实例
	Instance string `json:"instance,omitempty" xml:"instance,omitempty" form:"instance,omitempty" cbor:"instance,omitempty" yaml:"instance,omitempty" comment:"problem instance"`

	// 原始的错误状态码
	//
	// 根据定义此值是可选的字段，但是服务端发放的是始终有此值的。
	Status int `json:"status,omitempty" xml:"status,omitempty" form:"status,omitempty" cbor:"status,omitempty" yaml:"status,omitempty" comment:"problem status"`

	// 用户提交对象各个字段的错误信息
	Params []Param `json:"params,omitempty" xml:"params>i,omitempty" form:"params,omitempty" cbor:"params,omitempty" yaml:"params,omitempty" comment:"problem params"`

	// 反馈给用户的信息
	Extensions any `json:"extensions,omitempty" xml:"extensions,omitempty" form:"extensions,omitempty" cbor:"extensions,omitempty" yaml:"extensions,omitempty" comment:"problem extensions"`

	// 是否已经通过 [Problem.Apply] 放回对象池中。
	// Apply 是公开对象，如果多次调用可能会被放回多次，
	// 此标记可以记录是否已经放回对象池。
	isFree bool
	ps     *Problems[T]
}

// Param 单个错误字段的描述
type Param struct {
	Name    string   `json:"name" xml:"name" form:"name" cbor:"name" yaml:"name"`                // 出错字段的名称
	Reasons []string `json:"reasons" xml:"reasons" form:"reasons" cbor:"reasons" yaml:"reasons"` // 出错信息
}

// New 创建一个 [Problem] 对象
func (ps *Problems[T]) New() *Problem[T] {
	p := ps.problemPool.Get().(*Problem[T])
	if p.Params != nil {
		p.Params = p.Params[:0]
	}
	p.Extensions = nil
	p.Instance = ""
	p.Detail = ""
	p.Type = ""
	p.Title = ""
	p.Status = 0
	p.isFree = false
	p.ps = ps

	return p
}

// Init 初始化 p
func (ps *Problems[T]) Init(p *Problem[T], id ID, printer *localeutil.Printer) *Problem[T] {
	item := ps.Get(id)
	p.Type = item.Type
	p.Title = item.Title.LocaleString(printer)
	p.Status = item.Status
	if item.Detail != nil {
		p.Detail = item.Detail.LocaleString(printer)
	}
	return p
}

// MarshalHTML 实现 [negotiation/html.Marshaler] 接口
func (p *Problem[T]) MarshalHTML() (string, any) { return "problem", p }

func (p *Problem[T]) Apply(ctx T) {
	ctx.Render(p.Status, p)
	p.ps.Free(p)
}

// Free 回收 [Problem] 对象
func (ps *Problems[T]) Free(p *Problem[T]) {
	if len(p.Params) < 30 && !p.isFree {
		p.isFree = true
		p.ps.problemPool.Put(p)
	}
}

// WithParam 添加具体的错误字段及描述信息
func (p *Problem[T]) WithParam(name, reason string) *Problem[T] {
	if idx := slices.IndexFunc(p.Params, func(pp Param) bool { return pp.Name == name }); idx > -1 {
		param := p.Params[idx]
		param.Reasons = append(param.Reasons, reason)
		p.Params[idx] = param
	} else {
		p.Params = append(p.Params, Param{Name: name, Reasons: []string{reason}})
	}

	return p
}

// WithExtensions 指定扩展对象信息
//
// 多次调用将会覆盖之前的内容。
func (p *Problem[T]) WithExtensions(ext any) *Problem[T] {
	p.Extensions = ext
	return p
}

// WithInstance 指定发生错误的实例
//
// 多次调用将会覆盖之前的内容。默认为 [Context.ID]。
func (p *Problem[T]) WithInstance(instance string) *Problem[T] {
	p.Instance = instance
	return p
}
