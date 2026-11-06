// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"fmt"
	"iter"
	"slices"
	"sync"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/internal/status"
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

	Problems struct {
		problems []*LocaleProblem // 需保证元素的顺序相同
	}

	LocaleProblem struct {
		Type          string
		Title, Detail LocaleStringer

		status int
	}

	problemError struct {
		err error
		id  string
	}
)

func newProblem() *Problem {
	p := problemPool.Get().(*Problem)
	if p.Params != nil {
		p.Params = p.Params[:0]
	}
	p.Extensions = nil
	p.Instance = ""
	return p
	// 其它的基本字段在 [Problems.initProblem] 中初始化
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
func (ctx *Context) Problem(id string) *Problem { return ctx.initProblem(newProblem(), id) }

// ProblemFromStatus 从 HTTP 状态码生成 [Problem]
func (ctx *Context) ProblemFromStatus(status int) *Problem {
	if id, found := problemsID[status]; found {
		return ctx.Problem(id)
	}

	return ctx.Log(NewError("Not found problem from status %d", status), ProblemInternalServerError)
}

func (ctx *Context) initProblem(pp *Problem, id string) *Problem {
	ctx.Server().Problems().initProblem(pp, id, ctx.LocalePrinter())
	return pp.WithInstance(ctx.ID())
}

// Log 记录日志并根据情况生成对应的 [Problem] 对象
//
// problemID 表示生成的 [Problem] 对象的 Type 值，如果 err 本身就是 [Problem] 类型，
// 那么将忽略 problemID 值，若是两者都不满足，则将 problemID 设置为 [ProblemInternalServerError]。
//
// problemID 同时也是日志分类的依据，比如所有 4xx 的日志会写入 WARN，其它的才会写入 ERROR。
func (ctx *Context) Log(err error, problemID string) *Problem {
	pe, ok := errors.AsType[*problemError](err)
	if ok {
		problemID = pe.id
	} else {
		if problemID == "" {
			problemID = ProblemInternalServerError
		}
	}

	perr := ctx.Problem(problemID)

	if perr.Status < 500 {
		ctx.Logs().WARN().Handler().Handle(ctx.Logs().NewRecord().DepthError(3, err))
	} else {
		ctx.Logs().ERROR().Handler().Handle(ctx.Logs().NewRecord().DepthError(3, err))
	}

	return perr
}

func (ctx *Context) NotFound() *Problem { return ctx.Problem(ProblemNotFound) }

func (ctx *Context) NotImplemented() *Problem { return ctx.Problem(ProblemNotImplemented) }

// newProblems 创建用于管理 [Problem] 的对象
func newProblems() *Problems {
	ps := &Problems{
		problems: make([]*LocaleProblem, 0, 100),
	}
	initProblems(ps)
	return ps
}

// Problems 提供 [Problem] 管理
func (s *Server) Problems() *Problems { return s.problems }

// Add 添加新项
//
// NOTE: 已添加的内容无法修改，如果确实有需求，只能通过修改翻译项的方式间接进行修改。
func (ps *Problems) Add(s int, p ...*LocaleProblem) *Problems {
	if !status.IsProblemStatus(s) { // 只需验证大于 400 的状态码。
		panic("status 必须是一个有效的状态码")
	}

	for _, pp := range p {
		if ps.Exists(pp.Type) {
			panic(fmt.Sprintf("存在相同值的 Type 参数 %s", pp.Type))
		}

		if pp.Title == nil {
			panic("title 不能为空")
		}

		if pp.Detail == nil {
			panic("detail 不能为空")
		}

		pp.status = s

		ps.problems = append(ps.problems, pp)
	}

	return ps
}

// Exists 查看指定 id 是否已经存在
func (ps *Problems) Exists(id string) bool {
	return slices.IndexFunc(ps.problems, func(p *LocaleProblem) bool { return p.Type == id }) > -1
}

// Problems 所有已注册的错误代码
func (ps *Problems) Problems() iter.Seq2[int, *LocaleProblem] {
	return func(yield func(int, *LocaleProblem) bool) {
		for _, s := range ps.problems {
			if !yield(s.status, s) {
				break
			}
		}
	}
}

func (ps *Problems) initProblem(pp *Problem, id string, p *localeutil.Printer) {
	if i := slices.IndexFunc(ps.problems, func(p *LocaleProblem) bool { return p.Type == id }); i > -1 {
		sp := ps.problems[i]
		pp.Type = sp.Type
		pp.Title = sp.Title.LocaleString(p)
		pp.Detail = sp.Detail.LocaleString(p)
		pp.Status = sp.status
		return
	}
	panic(fmt.Sprintf("未找到有关 %s 的定义", id)) // 初始化时没有给定相关的定义，所以直接 panic。
}

// IsProblem 这是表示错误的状态码
func IsProblem(status int) bool {
	_, f := problemsID[status]
	return f
}

// NewProblemError 将 err 以指定的 problemID 进行包装
//
// 当前框架在处理此类错误时，在记录错误的同时，会将 problemID 值反馈给客户端。
func NewProblemError(problemID string, err error) error {
	if err == nil {
		panic("参数 err 不能为空")
	}

	return &problemError{id: problemID, err: err}
}

func NewProblemErrorFromStatus(status int, err error) error {
	if id, found := problemsID[status]; found {
		return NewProblemError(id, err)
	}

	// 将 err 用 problemError 包裹的目的主要是为显示一个特定的状态码，
	// 如果找不到，仅对客户端的状态码有影响，但是 err 本身没有问题。
	return err
}

func (err *problemError) Unwrap() error { return err.err }

func (err *problemError) Error() string { return err.err.Error() }
