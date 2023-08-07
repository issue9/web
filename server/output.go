// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"net/http"

	"github.com/issue9/localeutil"
	"golang.org/x/text/transform"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/problems"
)

// Responser 向客户端输出对象需要实现的接口
type Responser interface {
	// Apply 通过 [Context] 将当前内容渲染到客户端
	//
	// 在调用 Apply 之后，就不再使用 [Responser] 对象。
	// 如果你的对象支持 [sync.Pool] 的复用方式，可以在此方法中回收内存。
	Apply(*Context)
}

type ResponserFunc func(*Context)

func (f ResponserFunc) Apply(c *Context) { f(c) }

// SetWriter 自定义输出通道
//
// f 用于构建一个用于输出的 [http.ResponseWriter] 接口对象，其原型为：
//
//	func(w http.ResponseWriter) http.ResponseWriter
//
// 其中 w 表示原本与 [Context] 关联的对象，用户可以基于此对象作二次封装，
// 或是完全舍弃，都是可以的。
//
// 如果已经有内容输出，此操作将会 panic。
func (ctx *Context) SetWriter(f func(http.ResponseWriter) http.ResponseWriter) {
	if ctx.Wrote() {
		panic("已有内容输出，不可再更改！")
	}
	if f == nil {
		panic("参数 w 不能为空")
	}

	resp := f(ctx.originResponse)
	ctx.originResponse = resp
	ctx.writer = resp
}

// Render 向客户端输出对象
//
// status 想输出给用户状态码，如果出错，那么最终展示给用户的状态码可能不是此值；
// body 表示输出的对象，该对象最终调用 [Context.Marshal] 编码；
// problem 表示 body 是否为 [Problem] 对象，对于 Problem 对象可能会有特殊的处理；
func (ctx *Context) Render(status int, body any, problem bool) {
	// NOTE: 此方法不返回错误代码，所有错误在方法内直接处理。
	// 输出对象时若出错，状态码也已经输出，此时向调用方报告错误，
	// 调用方除了输出错误日志，也没有其它面向客户的补救措施。

	if body == nil {
		ctx.WriteHeader(status)
		return
	}

	ctx.Header().Set(header.ContentType, header.BuildContentType(ctx.Mimetype(problem), ctx.Charset()))
	if id := ctx.LanguageTag().String(); id != "" {
		ctx.Header().Set(header.ContentLang, id)
	}

	data, err := ctx.Marshal(body)
	if err != nil {
		ctx.Logs().ERROR().Printf("%+v", err)

		if problem {
			ctx.WriteHeader(status)
		} else {
			id := problems.ProblemNotAcceptable
			ctx.Render(problems.Status(id), ctx.Problem(id), true)
		}
		return
	}

	ctx.WriteHeader(status)
	if _, err = ctx.Write(data); err != nil {
		ctx.Logs().ERROR().Printf("%+v", err)
	}
}

// Marshal 将对象 v 按用户要求编码并返回
func (ctx *Context) Marshal(v any) ([]byte, error) {
	if ctx.outputMimetype.Marshal == nil { // 该值是可以为 nil 的，比如上传等操作。
		return nil, localeutil.Error("not found serialization for %s", ctx.Mimetype(false))
	}
	return ctx.outputMimetype.Marshal(ctx, v)
}

// Wrote 是否已经有内容输出
func (ctx *Context) Wrote() bool { return ctx.wrote }

// Sprintf 将内容翻译成当前请求的语言
func (ctx *Context) Sprintf(key string, v ...any) string {
	return ctx.LocalePrinter().Sprintf(key, v...)
}

// Write 向客户端输出内容
//
// 如非必要，应该返回 [Responser] 进行输出。
func (ctx *Context) Write(bs []byte) (int, error) {
	if len(bs) == 0 {
		return 0, nil
	}

	if !ctx.Wrote() { // 在第一次有内容输出时，才决定构建 Encoding 和 Charset 的 io.Writer
		ctx.wrote = true

		if ctx.outputEncoding != nil {
			ctx.encodingCloser = ctx.outputEncoding.Get(ctx.writer)
			ctx.writer = ctx.encodingCloser
		}

		if !header.CharsetIsNop(ctx.outputCharset) {
			ctx.charsetCloser = transform.NewWriter(ctx.writer, ctx.outputCharset.NewEncoder())
			ctx.writer = ctx.charsetCloser
		}
	}

	if ctx.status < 199 {
		ctx.WriteHeader(http.StatusOK)
	}
	return ctx.writer.Write(bs)
}

// WriteHeader 向客户端输出 HTTP 代码
//
// 如非必要，应该返回 [Responser] 进行输出。
func (ctx *Context) WriteHeader(status int) {
	if ctx.status > 199 && ctx.status != status {
		panic(fmt.Sprintf("已有状态码 %d，再次设置无效 %d", ctx.status, status))
	}

	ctx.Header().Del(header.ContentLength) // https://github.com/golang/go/issues/14975
	ctx.status = status
	ctx.originResponse.WriteHeader(status)
}

func (ctx *Context) Header() http.Header { return ctx.originResponse.Header() }
