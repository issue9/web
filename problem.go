// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"unique"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/problems"
)

type (
	Problem = problems.Problem[*Context]

	problemError struct {
		err error
		id  unique.Handle[string]
	}
)

// Problem 返回指定 id 的 [Problem]
func (ctx *Context) Problem(id unique.Handle[string]) *Problem {
	return ctx.initProblem(ctx.Server().Problems().NewProblem(), id)
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
func (s *Server) Problems() *problems.Problems[*Context] { return s.problems }

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
