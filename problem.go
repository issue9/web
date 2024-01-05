// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"sync"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/status"
)

const rfc7807PoolMaxParams = 30 // len(RFC7807.Params) 少于此值才会回收。

var rfc7807Pool = &sync.Pool{New: func() any { return &RFC7807{} }}

type (
	// Problem 向用户反馈非正常信息的对象接口
	Problem interface {
		Responser

		// WithParam 添加具体的错误字段及描述信息
		//
		// 如果已经存在同名，则会 panic。
		WithParam(name, reason string) Problem

		// WithExtensions 指定扩展对象信息
		//
		// 多次调用将会覆盖之前的内容。
		WithExtensions(any) Problem

		// WithInstance 指定发生错误的实例
		//
		// 多次调用将会覆盖之前的内容。默认为 [Context.ID]。
		WithInstance(string) Problem

		// 不允许其它实现
		private()
	}

	// RFC7807 [Problem] 的 [RFC7807] 实现
	//
	// [MarshalFunc] 的实现者，可能需要对 [RFC7807] 进行处理以便输出更加友好的格式。
	//
	// [RFC7807]: https://datatracker.ietf.org/doc/html/RFC7807
	RFC7807 struct {
		// NOTE: 无法缓存内容，因为用户请求的语言每次都可能是不一样的。
		// NOTE: Problem 应该是 final 状态的，否则像 [Context.PathID] 等的实现需要指定 Problem 对象。
		// NOTE: 这是 [Problem] 接口的唯一实现，之所以多此一举用，是因为像 [FilterContext.Problem]
		// 的返回值如果是 [RFC7807] 而不是 [Problem]，那么在 [HandleFunc] 中作为 [Responser] 返回时需要再次判断是否为 nil。

		Type       string         `json:"type" xml:"type" form:"type"`
		Title      string         `json:"title" xml:"title" form:"title"`
		Detail     string         `json:"detail,omitempty" xml:"detail,omitempty" form:"detail,omitempty"`
		Instance   string         `json:"instance,omitempty" xml:"instance,omitempty" form:"instance,omitempty"`
		Status     int            `json:"status" xml:"status" form:"status"`
		Extensions any            `json:"extensions,omitempty" xml:"extensions,omitempty" form:"extensions,omitempty"` // 反馈给用户的信息
		Params     []RFC7807Param `json:"params,omitempty" xml:"params>i,omitempty" form:"params,omitempty"`           // 用户提交对象各个字段的错误信息
	}

	RFC7807Param struct {
		Name   string `json:"name" xml:"name" form:"name"`       // 出错字段的名称
		Reason string `json:"reason" xml:"reason" form:"reason"` // 出错信息
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

func newRFC7807() *RFC7807 {
	p := rfc7807Pool.Get().(*RFC7807)
	if p.Params != nil {
		p.Params = p.Params[:0]
	}
	p.Extensions = nil
	p.Instance = ""
	return p
	// 其它的基本字段在 [Problems.initProblem] 中初始化
}

func (p *RFC7807) Error() string { return p.Title }

func (p *RFC7807) Apply(ctx *Context) Problem {
	// NOTE: 此方法要始终返回 nil

	ctx.Header().Set(header.ContentType, header.BuildContentType(ctx.Mimetype(true), ctx.Charset()))
	if id := ctx.LanguageTag().String(); id != "" {
		ctx.Header().Set(header.ContentLang, id)
	}

	ctx.WriteHeader(p.Status) // 调用之后，报头不再启作用

	data, err := ctx.Marshal(p)
	if err != nil {
		ctx.Logs().ERROR().Error(err)
		return nil
	}

	if _, err = ctx.Write(data); err != nil {
		ctx.Logs().ERROR().Error(err)
	}
	if len(p.Params) < rfc7807PoolMaxParams {
		rfc7807Pool.Put(p)
	}

	return nil
}

func (p *RFC7807) WithParam(name, reason string) Problem {
	if slices.IndexFunc(p.Params, func(pp RFC7807Param) bool { return pp.Name == name }) > -1 {
		panic("已经存在")
	}
	p.Params = append(p.Params, RFC7807Param{Name: name, Reason: reason})
	return p
}

func (p *RFC7807) WithExtensions(ext any) Problem {
	p.Extensions = ext
	return p
}

func (p *RFC7807) WithInstance(instance string) Problem {
	p.Instance = instance
	return p
}

func (p *RFC7807) private() {}

// Problem 返回指定 id 的 [Problem]
func (ctx *Context) Problem(id string) Problem { return ctx.initProblem(newRFC7807(), id) }

func (ctx *Context) initProblem(pp *RFC7807, id string) Problem {
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
func (ctx *Context) Error(err error, problemID string) Problem {
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

func (ctx *Context) NotFound() Problem { return ctx.Problem(ProblemNotFound) }

func (ctx *Context) NotImplemented() Problem { return ctx.Problem(ProblemNotImplemented) }

// Problem 如果有错误信息转换成 [Problem] 否则返回 nil
func (v *FilterContext) Problem(id string) Problem {
	if v == nil || v.len() == 0 {
		return nil
	}
	return v.Context().initProblem(v.problem, id)
}

func InternalNewProblems(prefix string) *Problems {
	ps := &Problems{
		prefix:   prefix,
		problems: make([]*LocaleProblem, 0, 100),
	}
	initProblems(ps)

	return ps
}

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
func (ps *Problems) Visit(visit func(status int, p *LocaleProblem)) {
	for _, s := range ps.problems {
		visit(s.status, s)
	}
}

func (ps *Problems) initProblem(pp *RFC7807, id string, p *localeutil.Printer) {
	if i := slices.IndexFunc(ps.problems, func(p *LocaleProblem) bool { return p.ID == id }); i > -1 {
		sp := ps.problems[i]
		pp.Type = sp.typ
		pp.Title = sp.Title.LocaleString(p)
		pp.Detail = sp.Detail.LocaleString(p)
		pp.Status = sp.status
		return
	}
	panic(fmt.Sprintf("未找到有关 %s 的定义", id)) // 初始化时没有给定相关的定义，所以直接 panic。
}
