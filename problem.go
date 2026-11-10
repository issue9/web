// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"slices"
	"sync"
	"unique"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/problems"
)

var problemPool = &sync.Pool{New: func() any { return &Problem{} }}

type (
	// Problem 基于 [RFC7807] 和 [RFC9457] 用于描述错误信息的对象
	//
	// 这是一个提供给客户端的信息，不能将敏感信息放于其中。
	//
	// [RFC7807]: https://www.rfc-editor.org/info/rfc7807
	// [RFC9457]: https://www.rfc-editor.org/info/rfc9457
	Problem struct {
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
		Params []ProblemParam `json:"params,omitempty" xml:"params>i,omitempty" form:"params,omitempty" cbor:"params,omitempty" yaml:"params,omitempty" comment:"problem params"`

		// 反馈给用户的信息
		Extensions any `json:"extensions,omitempty" xml:"extensions,omitempty" form:"extensions,omitempty" cbor:"extensions,omitempty" yaml:"extensions,omitempty" comment:"problem extensions"`
	}

	// ProblemParam 单个错误字段的描述
	ProblemParam struct {
		Name   string `json:"name" xml:"name" form:"name" cbor:"name" yaml:"name"`           // 出错字段的名称
		Reason string `json:"reason" xml:"reason" form:"reason" cbor:"reason" yaml:"reason"` // 出错信息
	}

	problemError struct {
		err error
		id  unique.Handle[string]
	}
)

func newProblem() *Problem {
	p := problemPool.Get().(*Problem)
	if p.Params != nil {
		p.Params = p.Params[:0]
	}
	p.Extensions = nil
	p.Instance = ""
	p.Detail = ""
	p.Type = ""
	p.Title = ""
	p.Status = 0
	return p
}

// MarshalHTML 实现 [mimetype/html.Marshaler] 接口
func (p *Problem) MarshalHTML() (string, any) { return "problem", p }

func (p *Problem) Apply(ctx *Context) {
	ctx.Render(p.Status, p)
	if len(p.Params) < 30 {
		problemPool.Put(p)
	}
}

// WithParam 添加具体的错误字段及描述信息
//
// 如果已经存在同名，则会 panic。
func (p *Problem) WithParam(name, reason string) *Problem {
	if slices.IndexFunc(p.Params, func(pp ProblemParam) bool { return pp.Name == name }) > -1 {
		panic("已经存在")
	}
	p.Params = append(p.Params, ProblemParam{Name: name, Reason: reason})
	return p
}

// WithExtensions 指定扩展对象信息
//
// 多次调用将会覆盖之前的内容。
func (p *Problem) WithExtensions(ext any) *Problem {
	p.Extensions = ext
	return p
}

// WithInstance 指定发生错误的实例
//
// 多次调用将会覆盖之前的内容。默认为 [Context.ID]。
func (p *Problem) WithInstance(instance string) *Problem {
	p.Instance = instance
	return p
}

// Problem 返回指定 id 的 [Problem]
func (ctx *Context) Problem(id unique.Handle[string]) *Problem {
	return ctx.initProblem(newProblem(), id)
}

func (ctx *Context) initProblem(p *Problem, id unique.Handle[string]) *Problem {
	initProblem(ctx.Server().Problems().Get(id), p, ctx.LocalePrinter())
	return p.WithInstance(ctx.ID())
}

func initProblem(item *problems.Item, p *Problem, printer *localeutil.Printer) {
	p.Type = item.Type
	p.Title = item.Title.LocaleString(printer)
	p.Status = item.Status
	if item.Detail != nil {
		p.Detail = item.Detail.LocaleString(printer)
	}
}

// Log 记录日志并根据情况生成对应的 [Problem] 对象
//
// 如果 err 本身就是由 [NewProblemError] 创建的类型，
// 那么将忽略 problemID 值，直接使用 [NewProblemError] 中的 id。
// 若 problemID 为空则使用默认值 [problems.ProblemInternalServerError]。
//
// problemID 同时也是日志分类的依据，比如所有 4xx 的日志会写入 WARN，其它的才会写入 ERROR。
func (ctx *Context) Log(err error, problemID unique.Handle[string]) *Problem {
	if pe, ok := errors.AsType[*problemError](err); ok {
		problemID = pe.id
	} else {
		if problemID == problems.Empty {
			problemID = problems.InternalServerError
		}
	}

	p := ctx.Problem(problemID)

	if p.Status < 500 {
		ctx.Logs().WARN().Handler().Handle(ctx.Logs().NewRecord().DepthError(3, err))
	} else {
		ctx.Logs().ERROR().Handler().Handle(ctx.Logs().NewRecord().DepthError(3, err))
	}

	return p
}

func (ctx *Context) NotFound() *Problem { return ctx.Problem(problems.NotFound) }

func (ctx *Context) NotImplemented() *Problem { return ctx.Problem(problems.NotImplemented) }

// Problems 提供 [Problem] 管理
func (s *Server) Problems() *problems.Problems { return s.problems }

// NewProblemError 将 err 以指定的 problemID 进行包装
//
// 当前框架在处理此类错误时，在记录错误的同时，会将 problemID 值反馈给客户端。
func NewProblemError(problemID unique.Handle[string], err error) error {
	if err == nil {
		panic("参数 err 不能为空")
	}

	return &problemError{id: problemID, err: err}
}

func (err *problemError) Unwrap() error { return err.err }

func (err *problemError) Error() string { return err.err.Error() }
