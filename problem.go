// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"sync"

	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v8/header"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/internal/qheader"
	"github.com/issue9/web/internal/status"
)

var problemPool = &sync.Pool{New: func() any { return &Problem{} }}

type (
	// Problem 基于 [RFC7807] 用于描述错误信息的对象
	//
	// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
	Problem struct {
		XMLName  struct{} `xml:"problem" form:"-" cbor:"-" json:"-" html:"-"`
		Type     string   `json:"type" xml:"type" form:"type" cbor:"type"`
		Title    string   `json:"title" xml:"title" form:"title" cbor:"title"`
		Detail   string   `json:"detail,omitempty" xml:"detail,omitempty" form:"detail,omitempty" cbor:"detail,omitempty"`
		Instance string   `json:"instance,omitempty" xml:"instance,omitempty" form:"instance,omitempty" cbor:"instance,omitempty"`
		Status   int      `json:"status" xml:"status" form:"status" cbor:"status,omitempty"`

		// 用户提交对象各个字段的错误信息
		Params []ProblemParam `json:"params,omitempty" xml:"params>i,omitempty" form:"params,omitempty" cbor:"params,omitempty"`

		// 反馈给用户的信息
		Extensions any `json:"extensions,omitempty" xml:"extensions,omitempty" form:"extensions,omitempty" cbor:"extensions,omitempty"`
	}

	ProblemParam struct {
		Name   string `json:"name" xml:"name" form:"name" cbor:"name"`         // 出错字段的名称
		Reason string `json:"reason" xml:"reason" form:"reason" cbor:"reason"` // 出错信息
	}

	Problems struct {
		prefix   string
		problems []*LocaleProblem // 需保证元素的顺序相同
	}

	LocaleProblem struct {
		ID            string
		Title, Detail LocaleStringer

		status int
		typ    string // 带前缀的 ID 值
	}
)

// Type 相当于 RFC7807 的 type 属性
func (p *LocaleProblem) Type() string { return p.typ }

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

func (p *Problem) Error() string { return p.Title }

func (p *Problem) Apply(ctx *Context) {
	ctx.Header().Set(header.ContentType, qheader.BuildContentType(ctx.Mimetype(true), ctx.Charset()))
	if id := ctx.LanguageTag().String(); id != "" {
		ctx.Header().Set(header.ContentLanguage, id)
	}

	ctx.WriteHeader(p.Status) // Problem 先输出状态码

	data, err := ctx.Marshal(p)
	if err != nil {
		ctx.Logs().ERROR().Error(err)
		return
	}
	if _, err = ctx.Write(data); err != nil {
		ctx.Logs().ERROR().Error(err)
	}

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

// Error 将 err 输出到 ERROR 通道并尝试以指定 id 的 [Problem] 返回
//
// 如果 id 为空，尝试以下顺序获得值：
//   - err 是否是由 [NewError] 创建，如果是则采用 err.Status 取得 ID 值；
//   - err 是否为 [fs.ErrPermission]，如果是采用 [ProblemForbidden] 作为 ID；
//   - err 是否为 [fs.ErrNotExist]，如果是采用 [ProblemNotFound] 作为 ID；
//   - 采用 [ProblemInternalServerError]；
func (ctx *Context) Error(err error, problemID string) *Problem {
	if problemID == "" {
		var herr *errs.HTTP
		switch {
		case errors.As(err, &herr):
			problemID = problemsID[herr.Status]
			err = herr.Message
		case errors.Is(err, fs.ErrPermission):
			problemID = ProblemForbidden
		case errors.Is(err, fs.ErrNotExist):
			problemID = ProblemNotFound
		default:
			problemID = ProblemInternalServerError
		}
	}

	ctx.Logs().ERROR().Handler().Handle(ctx.Logs().NewRecord().DepthError(3, err))
	return ctx.Problem(problemID)
}

func (ctx *Context) NotFound() *Problem { return ctx.Problem(ProblemNotFound) }

func (ctx *Context) NotImplemented() *Problem { return ctx.Problem(ProblemNotImplemented) }

// Problem 如果有错误信息转换成 [Problem] 否则返回 nil
func (v *FilterContext) Problem(id string) Responser {
	if v == nil || v.len() == 0 {
		return nil
	}
	return v.Context().initProblem(v.problem, id)
}

func newProblems(prefix string) *Problems {
	ps := &Problems{
		prefix:   prefix,
		problems: make([]*LocaleProblem, 0, 100),
	}
	initProblems(ps)
	return ps
}

func (s *InternalServer) Problems() *Problems { return s.problems }

// Prefix 所有 ID 的统一前缀
func (ps *Problems) Prefix() string { return ps.prefix }

// Add 添加新项
//
// NOTE: 已添加的内容无法修改，如果确实有需求，只能通过修改翻译项的方式间接进行修改。
func (ps *Problems) Add(s int, p ...*LocaleProblem) *Problems {
	if !status.IsProblemStatus(s) { // 只需验证大于 400 的状态码。
		panic("status 必须是一个有效的状态码")
	}

	for _, pp := range p {
		if ps.exists(pp.ID) {
			panic(fmt.Sprintf("存在相同值的 id 参数 %s", pp.ID))
		}

		if pp.Title == nil {
			panic("title 不能为空")
		}

		if pp.Detail == nil {
			panic("detail 不能为空")
		}

		if ps.Prefix() == ProblemAboutBlank {
			pp.typ = ProblemAboutBlank
		} else {
			pp.typ = ps.Prefix() + pp.ID
		}

		pp.status = s

		ps.problems = append(ps.problems, pp)
	}

	return ps
}

func (ps *Problems) exists(id string) bool {
	return slices.IndexFunc(ps.problems, func(p *LocaleProblem) bool { return p.ID == id }) > -1
}

// Visit 遍历错误代码
//
// visit 签名：
//
//	func(status int, p *LocaleProblem)
//
// status 该错误代码反馈给用户的 HTTP 状态码；
func (ps *Problems) Visit(visit func(int, *LocaleProblem)) {
	for _, s := range ps.problems {
		visit(s.status, s)
	}
}

func (ps *Problems) initProblem(pp *Problem, id string, p *localeutil.Printer) {
	if i := slices.IndexFunc(ps.problems, func(p *LocaleProblem) bool { return p.ID == id }); i > -1 {
		sp := ps.problems[i]
		pp.Type = sp.Type()
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
