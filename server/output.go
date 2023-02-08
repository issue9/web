// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"golang.org/x/text/message"
	"golang.org/x/text/transform"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/problems"
)

var responsePool = &sync.Pool{New: func() any { return &response{} }}

type response struct {
	http.ResponseWriter
	w io.Writer
}

func (r *response) Write(data []byte) (int, error) { return r.w.Write(data) }

// SetWriter 将输出内容指向 w
//
// 如果已经有内容输出，此操作将会 panic。
func (ctx *Context) SetWriter(w io.Writer) {
	if ctx.Wrote() {
		panic("已有内容输出，不可再更改！")
	}
	if w == nil {
		panic("参数 w 不能为空")
	}

	resp := responsePool.Get().(*response)
	resp.ResponseWriter = ctx.resp
	resp.w = w
	ctx.resp = resp
	ctx.respWriter = resp
	ctx.OnExit(func(i int) { responsePool.Put(resp) })
}

// Marshal 向客户端输出内容
//
// status 想输出给用户状态码，如果出错，那么最终展示给用户的状态码可能不是此值；
// body 表示输出的对象，该对象最终调用 ctx.outputMimetype 编码；
// problem 表示 body 是否为 [Problem] 对象，对于 Problem 对象可能会有特殊的处理；
func (ctx *Context) Marshal(status int, body any, problem bool) {
	// NOTE: 此方法不返回错误代码，所有错误在方法内直接处理。
	// 输出对象时出错，状态码也已经输出，此时向调用方报告错误，
	// 调用方除了输出错误日志，也没有其它面向客户的补救措施。

	if ctx.status > 0 {
		panic("不能多次调用 Marshal 输出状态码")
	}

	if body == nil {
		ctx.WriteHeader(status)
		return
	}

	// 如果 outputMimetype.marshal 为空，说明在 Server.Mimetypes() 的配置中就是 nil。
	// 那么不应该执行到此，比如下载文件等直接从 ResponseWriter.Write 输出的。
	if ctx.outputMimetype.Marshal == nil {
		ctx.WriteHeader(http.StatusNotAcceptable)
		panic(fmt.Sprintf("未对 %s 作处理", ctx.Mimetype(false)))
	}

	ctx.Header().Set("Content-Type", header.BuildContentType(ctx.Mimetype(problem), ctx.Charset()))
	if id := ctx.LanguageTag().String(); id != "" {
		ctx.Header().Set("Content-Language", id)
	}

	data, err := ctx.outputMimetype.Marshal(ctx, body)
	if err != nil {
		ctx.Logs().ERROR().Printf("%+v", err)

		if problem {
			ctx.WriteHeader(status)
		} else {
			id := problems.ProblemNotAcceptable
			ctx.Marshal(problems.Status(id), ctx.Problem(id), true)
		}
		return
	}

	ctx.WriteHeader(status)
	if _, err = ctx.Write(data); err != nil {
		ctx.Logs().ERROR().Printf("%+v", err)
	}
}

// Wrote 是否已经有内容输出
func (ctx *Context) Wrote() bool { return ctx.wrote }

// Sprintf 将内容翻译成当前请求的语言
func (ctx *Context) Sprintf(key message.Reference, v ...any) string {
	return ctx.LocalePrinter().Sprintf(key, v...)
}

func (ctx *Context) Write(bs []byte) (int, error) {
	if !ctx.Wrote() { // 在第一次有内容输出时，才决定构建 Encoding 和 Charset 的 io.Writer
		ctx.wrote = true

		if ctx.outputEncoding != nil {
			ctx.encodingCloser = ctx.outputEncoding.Get(ctx.respWriter)
			ctx.respWriter = ctx.encodingCloser
		}

		if !header.CharsetIsNop(ctx.outputCharset) {
			ctx.charsetCloser = transform.NewWriter(ctx.respWriter, ctx.outputCharset.NewEncoder())
			ctx.respWriter = ctx.charsetCloser
		}
	}

	return ctx.respWriter.Write(bs)
}

func (ctx *Context) WriteHeader(status int) {
	ctx.Header().Del("Content-Length") // https://github.com/golang/go/issues/14975
	ctx.status = status
	ctx.resp.WriteHeader(status)
}

func (ctx *Context) Header() http.Header { return ctx.resp.Header() }
