// SPDX-License-Identifier: MIT

package server

import (
	"golang.org/x/text/message"

	"github.com/issue9/web/result"
)

// CTXSanitizer 提供对数据的验证和修正
//
// 但凡对象实现了该接口，那么在 Context.Read 和 Queries.Object
// 中会在解析数据成功之后，调用该接口进行数据验证。
type CTXSanitizer interface {
	CTXSanitize(*Context) result.Fields
}

// Result result.Result 与 Context 相结合的实现
type Result struct {
	result.Result
	ctx *Context
}

// ResultMessages 错误信息列表
//
// p 用于返回特定语言的内容。
func (srv *Server) ResultMessages(p *message.Printer) map[int]string {
	return srv.results.Messages(p)
}

// AddResultMessage 添加一条错误信息
//
// status 指定了该错误代码反馈给客户端的 HTTP 状态码；
// code 错误代码；
// key 和 v 表示错误信息的翻译项内容，最终会传递给 message.Printer.Sprintf 进行翻译；
func (srv *Server) AddResultMessage(status, code int, key message.Reference, v ...interface{}) {
	srv.results.AddMessage(status, code, key, v...)
}

// NewResult 返回 Result 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (ctx *Context) NewResult(code int) *Result {
	return ctx.newResult(ctx.server.results.NewResult(ctx.LocalePrinter, code))
}

// NewResultWithFields 返回 Result 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (ctx *Context) NewResultWithFields(code int, fields result.Fields) *Result {
	return ctx.newResult(ctx.server.results.NewResultWithFields(ctx.LocalePrinter, code, fields))
}

func (ctx *Context) newResult(rslt result.Result) *Result {
	return &Result{
		Result: rslt,
		ctx:    ctx,
	}
}

// Render 渲染内容
func (rslt *Result) Render() { rslt.ctx.Render(rslt.Status(), rslt.Result, nil) }

// RenderAndExit 渲染内容并退出当前的请求处理
func (rslt *Result) RenderAndExit() {
	rslt.Render()
	rslt.ctx.Exit(0)
}
