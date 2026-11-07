// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"fmt"
	"iter"
	"slices"
	"strings"
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

	Problems struct {
		problems []*LocaleProblem // 需保证元素的顺序相同
	}

	LocaleProblem struct {
		// ID 唯一 ID
		//
		// 与 Type 的区别在于 ID 用于查找，需要唯一性。
		// 而 Type 可能是 'about:blank' 或是不同的 URI。
		ID string

		Type          string
		Title, Detail LocaleStringer
		Status        int
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

	p := ctx.Problem(problemID)

	if p.Status < 500 {
		ctx.Logs().WARN().Handler().Handle(ctx.Logs().NewRecord().DepthError(3, err))
	} else {
		ctx.Logs().ERROR().Handler().Handle(ctx.Logs().NewRecord().DepthError(3, err))
	}

	return p
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

// Add 添加一组新项
//
// 在 p.Type 不是绝对 URL 的情况下，typePrefix 会自动加在其前部.做为前缀。
//
// NOTE: 已添加的内容无法修改，如果确实有需求，只能通过修改翻译项的方式间接进行修改。
func (ps *Problems) Add(typePrefix string, p ...*LocaleProblem) *Problems {
	for _, pp := range p {
		if pp.Type != aboutBlank && typePrefix != "" &&
			!strings.HasPrefix(pp.Type, "https://") && !strings.HasPrefix(pp.Type, "http://") {
			pp.Type = typePrefix + pp.Type
		}

		ps.add(pp)
	}
	return ps
}

func (ps *Problems) add(p *LocaleProblem) *Problems {
	if !status.IsProblemStatus(p.Status) { // 只需验证大于 400 的状态码。
		panic("status 必须是一个有效的状态码")
	}

	if p.ID == "" {
		panic("p.ID 不能为空")
	}

	if ps.Exists(p.ID) {
		panic(fmt.Sprintf("该 ID 值 %s 已经存在", p.ID))
	}

	if p.Type == "" {
		p.Type = aboutBlank
	}
	if p.Type != aboutBlank {
		if slices.ContainsFunc(ps.problems, func(pp *LocaleProblem) bool { return pp.Type == p.Type }) {
			panic(fmt.Sprintf("该 Type 值 %s 已经存在", p.Type))
		}
	}

	if p.Title == nil {
		panic("title 不能为空")
	}

	ps.problems = append(ps.problems, p)

	return ps
}

// Exists 查看指定 id 是否已经存在
func (ps *Problems) Exists(id string) bool {
	return slices.IndexFunc(ps.problems, func(p *LocaleProblem) bool { return p.ID == id }) > -1
}

// All 所有已注册的错误代码
func (ps *Problems) All() iter.Seq[*LocaleProblem] {
	return func(yield func(*LocaleProblem) bool) {
		for _, p := range ps.problems {
			if !yield(p) {
				break
			}
		}
	}
}

func (ps *Problems) initProblem(p *Problem, id string, printer *localeutil.Printer) {
	if i := slices.IndexFunc(ps.problems, func(lp *LocaleProblem) bool { return lp.ID == id }); i > -1 {
		lp := ps.problems[i]
		p.Type = lp.Type
		p.Title = lp.Title.LocaleString(printer)
		p.Status = lp.Status
		if lp.Detail != nil {
			p.Detail = lp.Detail.LocaleString(printer)
		}
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

func (err *problemError) Unwrap() error { return err.err }

func (err *problemError) Error() string { return err.err.Error() }
