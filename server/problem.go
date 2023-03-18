// SPDX-License-Identifier: MIT

package server

import (
	"sync"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/filter"
	"github.com/issue9/web/internal/problems"
	"github.com/issue9/web/logs"
)

const filterProblemPoolMaxSize = 20

var (
	rfc7807Pool = problems.NewRFC7807Pool[*Context]()

	filterProblemPool = &sync.Pool{New: func() any {
		const size = 5
		return &FilterProblem{
			keys:    make([]string, 0, size),
			reasons: make([]string, 0, size),
		}
	}}
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

		// AddParam 添加具体的错误字段及描述信息
		//
		// name 为字段名称；reason 为该字段的错误信息；
		// 可以多次添加相同 name 的项。
		AddParam(name string, reason string)
	}

	// BuildProblemFunc 生成 [Problem] 对象的方法
	//
	// id 表示当前错误信息的唯一值，该值有可能为 about:blank，表示不想向用户展示具体的值；
	// title 错误信息的简要描述；
	// status 输出的状态码；
	BuildProblemFunc func(id string, status int, title, detail string) Problem

	// FilterProblem 处理由过滤器生成的各错误
	FilterProblem struct {
		exitAtError bool
		ctx         *Context
		keys        []string
		reasons     []string
	}
)

// RFC7807Builder [BuildProblemFunc] 的 [RFC7807] 标准实现
//
// NOTE: 由于 www-form-urlencoded 对复杂对象的表现能力有限，
// 在此模式下将忽略由 [Problem.With] 添加的复杂类型，只保留基本类型。
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
func RFC7807Builder(id string, status int, title, detail string) Problem {
	return rfc7807Pool.New(id, status, title, detail)
}

// AddProblem 添加新的错误代码
func (srv *Server) AddProblem(id string, status int, title, detail localeutil.LocaleStringer) *Server {
	srv.problems.Add(id, status, title, detail)
	return srv
}

// VisitProblems 遍历错误代码
//
// visit 签名：
//
//	func(id string, status int, title, detail localeutil.LocaleStringer)
//
// id 为错误代码；
// status 该错误代码反馈给用户的 HTTP 状态码；
// title 错误代码的简要描述；
// detail 错误代码的明细；
func (srv *Server) VisitProblems(visit func(string, int, localeutil.LocaleStringer, localeutil.LocaleStringer)) {
	srv.problems.Visit(visit)
}

// Problem 返回批定 id 的错误信息
//
// id 通过此值从 [Problems] 中查找相应在的 title 并赋值给返回对象；
func (ctx *Context) Problem(id string) Problem {
	return ctx.Server().problems.Problem(ctx.LocalePrinter(), id)
}

// InternalServerError 输出 ERROR 通道并向返回 500 表示的 [Problem] 对象
func (ctx *Context) InternalServerError(err error) Problem {
	return ctx.logError(4, problems.ProblemInternalServerError, logs.Error, err)
}

// Error 将 err 输出到日志并以指定 id 的 [Problem] 返回
func (ctx *Context) Error(id string, level logs.Level, err error) Problem {
	return ctx.logError(4, id, level, err)
}

func (ctx *Context) logError(depth int, id string, level logs.Level, err error) Problem {
	ctx.Logs().NewEntry(level).DepthError(3, err)
	return ctx.Problem(id)
}

// NotFound 返回 id 值为 404 的 [Problem] 对象
func (ctx *Context) NotFound() Problem { return ctx.Problem(problems.ProblemNotFound) }

func (ctx *Context) NotImplemented() Problem { return ctx.Problem(problems.ProblemNotImplemented) }

// Logs 返回日志对象
//
// 所有日志接口都会根据 [Server.LocalePrinter] 进行本地化，规则如下：
//   - Logger.Error 如果参数实现了 localeutil.LocaleStringer 接口，会尝试本地化；
//   - Logger.String 会采用 [message.Printer.Sprintf] 进行本地化；
//   - Logger.Printf 会采用 [message.Printer.Sprintf] 进行本地化，且每个参数也将进行本地化；
//   - Logger.Print 对每个参数分别进行本地化，然后调用 [fmt.Sprint] 输出；
//   - Logger.Println 对每个参数分别进行本地化，然后调用 [fmt.Sprintln] 输出；
func (srv *Server) Logs() *logs.Logs { return srv.logs }

func (ctx *Context) NewFilterProblem(exitAtError bool) *FilterProblem {
	v := filterProblemPool.Get().(*FilterProblem)
	v.exitAtError = exitAtError
	v.keys = v.keys[:0]
	v.reasons = v.reasons[:0]
	v.ctx = ctx
	ctx.OnExit(func(*Context, int) {
		if len(v.keys) < filterProblemPoolMaxSize {
			filterProblemPool.Put(v)
		}
	})
	return v
}

func (v *FilterProblem) continueNext() bool { return !v.exitAtError || v.len() == 0 }

func (v *FilterProblem) len() int { return len(v.keys) }

// Add 直接添加一条错误信息
func (v *FilterProblem) Add(name string, reason localeutil.LocaleStringer) *FilterProblem {
	if v.continueNext() {
		return v.add(name, reason)
	}
	return v
}

// AddError 直接添加一条类型为 error 的错误信息
func (v *FilterProblem) AddError(name string, err error) *FilterProblem {
	if ls, ok := err.(localeutil.LocaleStringer); ok {
		return v.Add(name, ls)
	}
	return v.Add(name, localeutil.Phrase(err.Error()))
}

func (v *FilterProblem) add(name string, reason localeutil.LocaleStringer) *FilterProblem {
	v.keys = append(v.keys, name)
	v.reasons = append(v.reasons, reason.LocaleString(v.Context().LocalePrinter()))
	return v
}

func (v *FilterProblem) AddFilter(f filter.FilterFunc) *FilterProblem {
	if !v.continueNext() {
		return v
	}

	if name, msg := f(); msg != nil {
		v.add(name, msg)
	}
	return v
}

// When 只有满足 cond 才执行 f 中的验证
//
// f 中的 v 即为当前对象；
func (v *FilterProblem) When(cond bool, f func(v *FilterProblem)) *FilterProblem {
	if cond {
		f(v)
	}
	return v
}

// Context 返回关联的 [Context] 实例
func (v *FilterProblem) Context() *Context { return v.ctx }

// Problem 转换成 [Problem] 对象
//
// 如果当前对象没有收集到错误，那么将返回 nil。
func (v *FilterProblem) Problem(id string) Problem {
	if v == nil || v.len() == 0 {
		return nil
	}

	p := v.Context().Problem(id)
	for index, key := range v.keys {
		p.AddParam(key, v.reasons[index])
	}
	return p
}
