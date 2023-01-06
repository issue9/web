// SPDX-License-Identifier: MIT

package server

import (
	"github.com/issue9/logs/v4"
	"golang.org/x/text/message"

	"github.com/issue9/web/internal/problems"
)

type (
	// Problem API 错误信息对象需要实现的接口
	//
	// Problem 是对 [Responser] 细化，用于反馈给用户非正常状态下的数据，
	// 比如用户提交的数据错误，往往会返回 400 的状态码，
	// 并附带一些具体的字段错误信息，此类数据都可以以 Problem 对象的方式反馈给用户。
	//
	// 除了当前接口，该对象可能还要实现相应的序列化接口，比如要能被 JSON 解析，
	// 就要实现 json.Marshaler 接口或是相应的 struct tag。
	//
	// 并未规定实现者输出的字段名和布局，实现者可以根据 [BuildProblemFunc]
	// 给定的参数，结合自身需求决定。比如 [RFC7807Builder] 是对 [RFC7807] 的实现。
	//
	// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
	Problem interface {
		Responser

		// With 添加新的输出字段
		//
		// 如果添加的字段名称与现有的字段重名，应当 panic。
		With(key string, val any)

		// AddParam 添加数据验证错误信息
		AddParam(name string, reason string)
	}

	// BuildProblemFunc 生成 [Problem] 对象的方法
	//
	// id 表示当前错误信息的唯一值，该值有可能为 about:blank，表示不想向用户展示具体的值；
	// title 错误信息的简要描述；
	// status 输出的状态码；
	BuildProblemFunc func(id, title string, status int) Problem

	Problems interface {
		// TypePrefix [BuildProblemFunc] 参数 id 的前缀
		TypePrefix() string

		// SetTypePrefix 设置 id 的前缀
		//
		// 如果设置成 about:blank 那么将不输出 id
		SetTypePrefix(string)

		// Add 添加新的错误类型
		Add(...*StatusProblem)

		Exists(id string) bool

		Problems() []*StatusProblem

		// Problem 根据 id 生成 [Problem] 对象
		Problem(printer *message.Printer, id string) Problem

		// Status 添加一组固定状态码的错误码
		Status(int) *problems.StatusProblems[Problem]
	}

	StatusProblem = problems.StatusProblem

	// CTXSanitizer 在 [Context] 关联的上下文环境中对数据进行验证和修正
	//
	// 在 [Context.Read]、[Context.QueryObject] 以及 [Queries.Object]
	// 中会在解析数据成功之后会调用该接口。
	CTXSanitizer interface {
		// CTXSanitize 验证和修正当前对象的数据
		CTXSanitize(*Validation)
	}
)

var rfc7807Pool = problems.NewRFC7807Pool[*Context]()

// RFC7807Builder [BuildProblemFunc] 的 [RFC7807] 标准实现
//
// NOTE: 由于 www-form-urlencoded 对复杂对象的表现能力有限，
// 在此模式下将忽略由 [Problem.With] 添加的复杂类型，只保留基本类型。
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
func RFC7807Builder(id, title string, status int) Problem {
	return rfc7807Pool.New(id, title, status)
}

func (srv *Server) Problems() Problems { return srv.problems }

// Problem 转换成 [Problem] 对象
//
// 如果当前对象没有收集到错误，那么将返回 nil。
func (v *Validation) Problem(id string) Problem {
	if v == nil || v.Count() == 0 {
		return nil
	}

	p := v.Context().Problem(id)
	for index, key := range v.keys {
		p.AddParam(key, v.reasons[index])
	}
	return p
}

// Problem 返回批定 id 的错误信息
//
// id 通过此值从 [Problems] 中查找相应在的 title 并赋值给返回对象；
func (ctx *Context) Problem(id string) Problem {
	return ctx.Server().Problems().Problem(ctx.LocalePrinter(), id)
}

// InternalServerError 输出 ERROR 通道并向返回 500 表示的 [Problem] 对象
func (ctx *Context) InternalServerError(err error) Problem {
	return ctx.logError(3, "500", logs.LevelError, err)
}

// Error 将 err 输出到日志并以指定 id 的 [Problem] 返回
func (ctx *Context) Error(id string, level logs.Level, err error) Problem {
	return ctx.logError(3, id, level, err)
}

func (ctx *Context) logError(depth int, id string, level logs.Level, err error) Problem {
	entry := ctx.Logs().NewEntry(level).Location(depth)
	entry.Message = err.Error() // NOTE: 日志信息不会根据 ctx 作翻译
	ctx.Logs().Output(entry)
	return ctx.Problem(id)
}

// NotFound 返回 id 值为 404 的 [Problem] 对象
func (ctx *Context) NotFound() Problem { return ctx.Problem(problems.ProblemNotFound) }

func (ctx *Context) NotImplemented() Problem { return ctx.Problem(problems.ProblemNotImplemented) }
