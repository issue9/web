// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"unique"

	"github.com/issue9/web/problems"
)

type Problem = problems.Problem[*Context]

// Problem 返回指定 id 的 [Problem]
func (ctx *Context) Problem(id unique.Handle[string]) *Problem {
	ps := ctx.Server().Problems()
	p := ps.New()
	return ps.Init(p, id, ctx.LocalePrinter()).WithInstance(ctx.ID())
}

// Log 记录日志并根据情况生成对应的 [Problem] 对象
//
// 如果 err 本身就是由 [NewProblemError] 创建的类型，
// 那么将忽略 problemID 值，直接使用 [NewProblemError] 中的 id。
// 若 problemID 为空则使用默认值 [problems.ProblemInternalServerError]。
//
// problemID 同时也是日志分类的依据，比如所有 4xx 的日志会写入 WARN，其它的才会写入 ERROR。
func (ctx *Context) Log(err error, problemID unique.Handle[string]) *Problem {
	if pe, ok := errors.AsType[*problems.Error](err); ok {
		problemID = pe.ID()
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
