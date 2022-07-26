// SPDX-License-Identifier: MIT

package server

import (
	"sync"

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

	Problem struct {
		p problem.Problem
	}
)

// Problems [RFC7807] 错误代码管理
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
func (srv *Server) Problems() *problem.Problems { return srv.problems }

// With 添加新的字段名和值
func (p *Problem) With(key string, val any) *Problem {
	p.p.With(key, val)
	return p
}

func (p *Problem) AddParam(name string, reason ...string) *Problem {
	p.p.AddParam(name, reason...)
	return p
}

func (p *Problem) Apply(ctx *Context) {
	if err := ctx.Marshal(p.p.Status(), p.p); err != nil {
		ctx.Logs().ERROR().Error(err)
	}

	p.p.Destroy()
	problemPool.Put(p)
}

// Problem 向客户端输出错误信息
//
// id 通过此值从 [Problems] 中查找相应在的 title 和 detail 并赋值给返回对象；
// obj 表示实际的返回对象，如果为空，会采用 [problem.RFC7807]；
func (ctx *Context) Problem(id string, errs FieldErrs) *Problem {
	p := ctx.Server().Problems().Problem(id, ctx.LocalePrinter())
	for name, reason := range errs {
		p.AddParam(name, reason...)
	}

	pp := problemPool.Get().(*Problem)
	pp.p = p
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
