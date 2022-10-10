// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/text/message"
	"golang.org/x/text/transform"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/serializer"
)

type BeforeMarshalFunc func(*Context, any) any

type AfterMarshalFunc func(*Context, []byte) []byte

// OnMarshal 注册编码过程中的钩子函数
func (srv *Server) OnMarshal(mimetype string, before BeforeMarshalFunc, after AfterMarshalFunc) {
	if before != nil {
		if _, found := srv.beforeMarshals[mimetype]; found {
			panic(fmt.Sprintf("%s 已经存在", mimetype))
		}
		srv.beforeMarshals[mimetype] = before
	}

	if after != nil {
		if _, found := srv.afterMarshals[mimetype]; found {
			panic(fmt.Sprintf("%s 已经存在", mimetype))
		}
		srv.afterMarshals[mimetype] = after
	}
}

// Marshal 向客户端输出内容
//
// status 想输出给用户状态码，如果出错，那么最终展示给用户的状态码可能不是此值；
// body 表示输出的对象，该对象最终调用 ctx.outputMimetype 编码；
// problem 表示 body 是否为 [Problem] 对象，对于 Problem 对象可能会有特殊的处理；
func (ctx *Context) Marshal(status int, body any, problem bool) error {
	if body == nil {
		ctx.WriteHeader(status)
		return nil
	}

	// 如果 outputMimetype 为空，说明在 Server.Mimetypes() 的配置中就是 nil。
	// 那么不应该执行到此，比如下载文件等直接从 ResponseWriter.Write 输出的。
	if ctx.outputMimetype == nil {
		ctx.WriteHeader(http.StatusNotAcceptable)
		panic(fmt.Sprintf("未对 %s 作处理", ctx.outputMimetypeName))
	}

	if problem {
		ctx.outputMimetypeName = ctx.Server().Problems().mimetype(ctx.outputMimetypeName)
	}
	ctx.Header().Set("Content-Type", header.BuildContentType(ctx.outputMimetypeName, ctx.outputCharsetName))
	if id := ctx.languageTag.String(); id != "" {
		ctx.Header().Set("Content-Language", id)
	}

	if f, found := ctx.Server().beforeMarshals[ctx.outputMimetypeName]; found {
		body = f(ctx, body)
	}

	data, err := ctx.outputMimetype(body)
	switch {
	case err != nil && problem: // 如果在输出 problem 时出错状态码不变
		ctx.WriteHeader(status)
		return err
	case errors.Is(err, serializer.ErrUnsupported):
		ctx.WriteHeader(http.StatusNotAcceptable) // 此处不再输出 Problem 类型错误信息，大概率也是相同的错误。
		return err
	case err != nil:
		ctx.WriteHeader(http.StatusInternalServerError) // 此处不再输出 Problem 类型错误信息，大概率也是相同的错误。
		return err
	}

	if f, found := ctx.Server().afterMarshals[ctx.outputMimetypeName]; found {
		data = f(ctx, data)
	}

	ctx.WriteHeader(status)
	_, err = ctx.Write(data)
	return err
}

// Wrote 是否已经有内容输出
func (ctx *Context) Wrote() bool { return ctx.wrote }

// Sprintf 返回翻译后的结果
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
