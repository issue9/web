// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/issue9/mux/v9/header"
	"golang.org/x/text/transform"

	"github.com/issue9/web/internal/qheader"
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

// Wrap 替换底层的 [http.ResponseWriter] 对象
//
// f 用于构建新的 [http.ResponseWriter] 对象，其原型为：
//
//	func(w http.ResponseWriter) http.ResponseWriter
//
// 其中 w 表示原本与 [Context] 关联的对象，返回一个新的替换对象。
// 如果已经有内容输出，此操作将会 panic。
func (ctx *Context) Wrap(f func(http.ResponseWriter) http.ResponseWriter) {
	if ctx.Wrote() {
		panic("已有内容输出，不可再更改！")
	}
	if f == nil {
		panic("参数 f 不能为空")
	}

	resp := f(ctx.originResponse)
	ctx.originResponse = resp
	ctx.writer = resp
}

// Render 向客户端输出对象
//
// status 想输出给用户状态码，如果出错，那么最终展示给用户的状态码可能不是此值；
// body 表示输出的对象，该对象最终调用 [Context.Marshal] 编码；
func (ctx *Context) Render(status int, body any) {
	// NOTE: 此方法不返回错误代码，所有错误在方法内直接处理。输出对象时若出错，
	// 状态码也已经输出，此时向调用方报告错误，除了输出错误日志，也没有其它面向客户的补救措施。

	if ctx.s.onRender != nil {
		status, body = ctx.s.onRender(status, body)
	}

	if body == nil {
		ctx.WriteHeader(status)
		return
	}

	ctx.Header().Set(header.ContentType, qheader.BuildContentType(ctx.Mimetype(false), ctx.Charset()))
	if id := ctx.LanguageTag().String(); id != "" {
		ctx.Header().Set(header.ContentLanguage, id)
	}

	data, err := ctx.Marshal(body)
	if err != nil {
		// [Problem.Apply] 并未调用 [Context.Render]，应该不会死循环。
		var p *Problem
		if errors.As(err, &p) {
			p.Apply(ctx)
		} else {
			ctx.Error(err, ProblemNotAcceptable).Apply(ctx)
		}
		return
	}

	ctx.WriteHeader(status)
	if _, err = ctx.Write(data); err != nil {
		ctx.Logs().ERROR().Error(err)
	}
}

// Marshal 将对象 v 按用户要求编码并返回
func (ctx *Context) Marshal(v any) ([]byte, error) { return ctx.outputMimetype.Marshal(ctx, v) }

// Wrote 是否已经有内容输出
func (ctx *Context) Wrote() bool { return ctx.wrote }

// Sprintf 将内容翻译成当前请求的语言
func (ctx *Context) Sprintf(key string, v ...any) string {
	return ctx.LocalePrinter().Sprintf(key, v...)
}

// Write 向客户端输出内容
//
// NOTE: 如非必要，应该采用 [Context.Render] 输出。
func (ctx *Context) Write(bs []byte) (n int, err error) {
	if len(bs) == 0 {
		return 0, nil
	}

	if !ctx.Wrote() { // 在第一次有内容输出时，才决定构建 Compress 和 Charset 的 io.Writer
		ctx.wrote = true
		closes := make([]io.Closer, 0, 2)

		if ctx.outputCompressor != nil {
			w, err := ctx.outputCompressor.NewEncoder(ctx.writer)
			if err != nil {
				return 0, err
			}
			ctx.writer = w
			closes = append(closes, w)
		}

		if !qheader.CharsetIsNop(ctx.outputCharset) {
			ctx.Header().Add(header.Vary, header.ContentEncoding) // 只有在确定需要输出内容时才输出 Vary 报头
			w := transform.NewWriter(ctx.writer, ctx.outputCharset.NewEncoder())
			ctx.writer = w
			closes = append(closes, w)
		}

		if l := len(closes); l > 0 {
			if l > 1 {
				slices.Reverse(closes)
			}

			ctx.OnExit(func(*Context, int) {
				for _, c := range closes {
					if err := c.Close(); err != nil {
						ctx.Logs().ERROR().Error(err)
					}
				}
			})
		}
	}

	if ctx.status < http.StatusOK { // 1xx 可能还会改变状态码，比如 103
		ctx.WriteHeader(http.StatusOK)
	}
	return ctx.writer.Write(bs)
}

// WriteHeader 向客户端输出 HTTP 状态码
//
// NOTE: 如非必要，应该通过 [Context.Render] 输出。
func (ctx *Context) WriteHeader(status int) {
	if ctx.status >= http.StatusOK && ctx.status != status {
		panic(fmt.Sprintf("已有状态码 %d，再次设置无效 %d", ctx.status, status))
	}

	ctx.Header().Del(header.ContentLength) // https://github.com/golang/go/issues/14975
	ctx.status = status
	ctx.originResponse.WriteHeader(status)
}

func (ctx *Context) Header() http.Header { return ctx.originResponse.Header() }

// SetCookies 输出一组 Cookie
func (ctx *Context) SetCookies(c ...*http.Cookie) {
	for _, cookie := range c {
		http.SetCookie(ctx, cookie)
	}
}

// NotModified 决定何时可返回 304 状态码
//
// etag 返回当前内容关联的 ETag 报头内容，其原型为：
//
//	func()(etag string, weak bool)
//
// etag 表示对应的 etag 报头，需要包含双绰号，但是不需要 W/ 前缀，weak 是否为弱验证。
//
// body 获取返回给客户端的报文主体对象，其原型如下：
//
//	func()(body any, err error)
//
// 如果返回 body 的是 []byte 类型，会原样输出，其它类型则按照 [Context.Marshal] 进行转换成 []byte 之后输出。
// body 可能参与了 etag 的计算，为了防止重复生成 body，所以此函数允许 body 直接为 []byte 类型，
// 在不需要 body 参与计算 etag 的情况下，应该返回非 []byte 类型的 body。
func NotModified(etag func() (string, bool), body func() (any, error)) Responser {
	return ResponserFunc(func(ctx *Context) {
		if ctx.Request().Method == http.MethodGet {
			if tag, weak := etag(); qheader.InitETag(ctx, ctx.Request(), tag, weak) {
				ctx.WriteHeader(http.StatusNotModified)
				return
			}
		}

		b, err := body()
		if err != nil {
			ctx.Error(err, ProblemInternalServerError).Apply(ctx)
			return
		}

		if data, ok := b.([]byte); ok {
			ctx.Header().Set(header.ContentType, qheader.BuildContentType(ctx.Mimetype(false), ctx.Charset()))
			if id := ctx.LanguageTag().String(); id != "" {
				ctx.Header().Set(header.ContentLanguage, id)
			}

			ctx.WriteHeader(http.StatusOK)
			if _, err := ctx.Write(data); err != nil {
				ctx.Logs().ERROR().Error(err)
			}
		} else {
			ctx.Render(http.StatusOK, b)
		}
	})
}

// Status 仅向客户端输出状态码和报头
//
// kv 为报头，必须以偶数数量出现，奇数位为报头名，偶数位为对应的报头值；
//
// NOTE: 即使 code 为 400 等错误代码，当前函数也不会返回 [Problem] 对象。
func Status(code int, kv ...string) Responser {
	l := len(kv)
	if l > 0 && l%2 != 0 {
		panic("kv 必须偶数位")
	}

	return ResponserFunc(func(ctx *Context) {
		for i := 0; i < l; i += 2 {
			ctx.Header().Add(kv[i], kv[i+1])
		}
		ctx.WriteHeader(code)
	})
}

// Response 输出状态和对象至客户端
//
// body 表示需要输出的对象，该对象最终会被转换成相应的编码；
// kv 为报头，必须以偶数数量出现，奇数位为报头名，偶数位为对应的报头值；
func Response(status int, body any, kv ...string) Responser {
	l := len(kv)
	if l > 0 && l%2 != 0 {
		panic("kv 必须偶数位")
	}

	return ResponserFunc(func(ctx *Context) {
		for i := 0; i < l; i += 2 {
			ctx.Header().Add(kv[i], kv[i+1])
		}
		ctx.Render(status, body)
	})
}

func Created(v any, location string) Responser {
	if location != "" {
		return Response(http.StatusCreated, v, header.Location, location)
	}
	return Response(http.StatusCreated, v)
}

// OK 返回 200 状态码下的对象
func OK(v any) Responser { return Response(http.StatusOK, v) }

func NoContent() Responser { return Status(http.StatusNoContent) }

// Redirect 重定向至新的 URL
func Redirect(status int, url string) Responser {
	return Status(status, header.Location, url)
}

// KeepAlive 保持当前会话不退出
func KeepAlive(ctx context.Context) Responser {
	return ResponserFunc(func(*Context) { <-ctx.Done() })
}
