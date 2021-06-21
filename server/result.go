// SPDX-License-Identifier: MIT

package server

import "github.com/issue9/web/content"

// CTXSanitizer 提供对数据的验证和修正
//
// 但凡对象实现了该接口，那么在 Context.Read 和 Queries.Object
// 中会在解析数据成功之后，调用该接口进行数据验证。
type CTXSanitizer interface {
	CTXSanitize(*Context) content.Fields
}

// Result content.Result 与 Context 相结合的实现
type Result struct {
	content.Result
	ctx *Context
}

// NewResult 返回 Result 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (ctx *Context) NewResult(code int) *Result {
	return ctx.newResult(ctx.server.content.NewResult(ctx.LocalePrinter, code))
}

// NewResultWithFields 返回 Result 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (ctx *Context) NewResultWithFields(code int, fields content.Fields) *Result {
	return ctx.newResult(ctx.server.content.NewResultWithFields(ctx.LocalePrinter, code, fields))
}

func (ctx *Context) newResult(rslt content.Result) *Result {
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
