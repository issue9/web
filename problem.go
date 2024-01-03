// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"io/fs"
	"sync"

	"github.com/issue9/sliceutil"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/internal/header"
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
)

func newRFC7807() *RFC7807 {
	p := rfc7807Pool.Get().(*RFC7807)

	if p.Params != nil {
		p.Params = p.Params[:0]
	}
	p.Extensions = nil
	p.Instance = ""

	// 其它字段在 init 会被初始化

	return p
}

func (p *RFC7807) Init(id, title, detail string, status int) {
	p.Type = id
	p.Title = title
	p.Detail = detail
	p.Status = status
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
	if sliceutil.Exists(p.Params, func(pp RFC7807Param, _ int) bool { return pp.Name == name }) {
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

func (ctx *Context) initProblem(p *RFC7807, id string) Problem {
	ctx.Server().Problems().Init(p, id, ctx.LocalePrinter())
	return p.WithInstance(ctx.ID())
}

// Error 将 err 输出到 ERROR 通道并尝试以指定 id 的 [Problem] 返回
//
// 如果 id 为空，尝试以下顺序获得值：
//   - err 是否是由 [NewError] 创建，如果是则采用 err.Status 取得 ID 值；
//   - err 是否为 [fs.ErrPermission]，如果是采用 [ProblemForbidden] 作为 ID；
//   - err 是否为 [fs.ErrNotExist]，如果是采用 [ProblemNotFound] 作为 ID；
//   - 采用 [ProblemInternalServerError]；
func (ctx *Context) Error(err error, id string) Problem {
	if id == "" {
		var herr *errs.HTTP
		switch {
		case errors.As(err, &herr):
			id = problemsID[herr.Status]
			err = herr.Message
		case errors.Is(err, fs.ErrPermission):
			id = ProblemForbidden
		case errors.Is(err, fs.ErrNotExist):
			id = ProblemNotFound
		default:
			id = ProblemInternalServerError
		}
	}

	ctx.Logs().ERROR().Handler().Handle(ctx.Logs().NewRecord().DepthError(3, err))
	return ctx.Problem(id)
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
