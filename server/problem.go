// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"sync"

	"github.com/issue9/localeutil"
	"github.com/issue9/validation"

	"github.com/issue9/web/problem"
)

var problemPool = &sync.Pool{New: func() any { return &Problem{} }}

type FieldErrs = problem.FieldErrs

// CTXSanitizer 提供对数据的验证和修正
//
// 在 Context.Read 和 Queries.Object 中会在解析数据成功之后，调用该接口进行数据验证。
type (
	CTXSanitizer interface {
		// CTXSanitize 验证和修正当前对象的数据
		//
		// 如果验证有误，则需要返回这些错误信息。
		CTXSanitize(*Context) FieldErrs
	}

	Problems struct {
		typeBaseURL string
		problems    map[string]*statusProblem
		blank       bool // 在输出时所有的 Type 强制为 about:blank
	}

	statusProblem struct {
		status        int
		title, detail localeutil.LocaleStringer
	}

	Problem struct {
		status int
		ps     *Problems
		p      problem.Problem
	}
)

func newProblems() *Problems { return &Problems{problems: make(map[string]*statusProblem, 50)} }

// Problems [RFC7807] 错误代码管理
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
func (srv *Server) Problems() *Problems { return srv.problems }

// SetTypeBaseURL 设置 type 字段的基地址
func (p *Problems) SetTypeBaseURL(base string) { p.typeBaseURL = base }

// DisableType 禁止输出 type 字段
//
// 如果设置了此值，那么在输出内容时，所有的 type 字段会变成 about:blank。
// 此方法只影响由 [Problems.Problem] 生成的对象。
func (p *Problems) DisableType() { p.blank = true }

// AddProblem 添加新的错误类型
//
// id 表示 RFC7807 中的 type 值，要求必须唯一，且不能为空和 about:blank。
// [Problems.Problem] 会根据此值查找相应的文字说明给予 title 和 detail 字段；
// status 表示输出给客户端的状态码；
// title 和 detail 表示此 type 关联的标题和详细说明，
// 这些值有可能会赋予通过 [Problems.Problem] 生成的对象；
func (p *Problems) Add(id string, status int, title, detail localeutil.LocaleStringer) {
	if _, found := p.problems[id]; found {
		panic("存在相同值的 id 参数")
	}
	p.problems[id] = &statusProblem{status: status, title: title, detail: detail}
}

// Visit 遍历所有 Add 添加的项
//
// f 为遍历的函数，其原型为：
//  func(id string, status int, title, detail localeutil.LocaleStringer)
// 分别对应 [Problems.Add] 添加时的各个参数。
//
// 用户可以通过此方法生成 QA 页面。
func (p *Problems) Visit(f func(id string, status int, title, detail localeutil.LocaleStringer) bool) {
	for t, item := range p.problems {
		if !f(t, item.status, item.title, item.detail) {
			return
		}
	}
}

// WithTitle 修改 title 字段内容
//
// 如果不调用此方法，那么会继承由 [Problems.Add] 添加时的 title 值。
func (p *Problem) WithTitle(t string) *Problem {
	p.p.SetTitle(t)
	return p
}

// WithDetail 修改 detail 字段内容
//
// 如果不调用此方法，那么会继承由 [Problems.Add] 添加时的 detail 值。
func (p *Problem) WithDetail(d string) *Problem {
	p.p.SetDetail(d)
	return p
}

// WithStatus 指定输出的状态码
//
// 如果不调用此方法，那么会继承由 [Problems.Add] 添加时的 status 值。
func (p *Problem) WithStatus(s int) *Problem {
	p.p.SetStatus(s)
	return p
}

// WithInstance 指定 instance 字段
func (p *Problem) WithInstance(i string) *Problem {
	p.p.SetInstance(i)
	return p
}

func (p *Problem) Apply(ctx *Context) {
	if err := ctx.Marshal(p.status, p.p); err != nil {
		ctx.Logs().ERROR().Error(err)
	}

	p.p.Destroy()
	problemPool.Put(p)
}

// Problem 向客户端输出错误信息
//
// id 通过此值从 [Problems] 中查找相应在的 title 和 detail 并赋值给返回对象；
// obj 表示实际的返回对象，如果为空，会采用 [problem.RFC7807]；
func (ctx *Context) Problem(id string, obj problem.Problem) *Problem {
	ps := ctx.Server().Problems()

	sp, found := ps.problems[id]
	if !found {
		panic(fmt.Sprintf("未找到有关 %s 的定义", id))
	}

	if obj == nil {
		obj = problem.NewRFC7807(nil)
	}
	if ps.blank {
		obj.SetType("about:blank")
	} else {
		obj.SetType(ps.typeBaseURL + id)
	}
	obj.SetTitle(sp.title.LocaleString(ctx.LocalePrinter()))
	obj.SetDetail(sp.detail.LocaleString(ctx.LocalePrinter()))
	obj.SetStatus(sp.status)

	pp := problemPool.Get().(*Problem)
	pp.status = sp.status
	pp.ps = ctx.Server().Problems()
	pp.p = obj
	return pp
}

// NewValidation 声明验证器
//
// 一般配合 CTXSanitizer 接口使用：
//
//  type User struct {
//      Name string
//      Age int
//  }
//
//  func(o *User) CTXSanitize(ctx* web.Context) web.FieldErrs {
//      v := ctx.NewValidation(10)
//      return v.NewField(o.Name, "name", validator.Required().Message("不能为空")).
//          NewField(o.Age, "age", validator.Min(18).Message("不能小于 18 岁")).
//          LocaleMessages(ctx.localePrinter())
//  }
//
// cap 表示为错误信息预分配的大小；
func (ctx *Context) NewValidation(cap int) *validation.Validation {
	return validation.New(validation.ContinueAtError, cap)
}
