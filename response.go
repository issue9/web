// SPDX-License-Identifier: MIT

package web

import (
	"net/http"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/problems"
)

// NotModified 决定何时可返回 304 状态码
//
// etag 返回当前内容关联的 ETag 报头内容，其原型为：
//
//	func()(etag string, weak bool)
//
// etag 表示对应的 etag 报头，需要包含双绰号，但是不需要 W/ 前缀，weak 是否为弱验证。
//
// body 获取返回给客户端的报文主体对象，
// 如果返回的是 []byte 类型，会原样输出，
// 其它类型则按照 [Context.Marshal] 进行转换成 []byte 之后输出。
func NotModified(etag func() (string, bool), body func() (any, error)) Responser {
	return ResponserFunc(func(ctx *Context) {
		if ctx.Request().Method == http.MethodGet {
			if tag, weak := etag(); header.InitETag(ctx, ctx.Request(), tag, weak) {
				ctx.WriteHeader(http.StatusNotModified)
				return
			}
		}

		b, err := body()
		if err != nil {
			ctx.Logs().ERROR().Error(err)
			ctx.Render(problems.Status(ProblemInternalServerError), ctx.Problem(ProblemInternalServerError), true)
			return
		}

		var data []byte
		if d, ok := b.([]byte); ok {
			data = d
		} else {
			data, err = ctx.Marshal(b)
			if err != nil {
				ctx.Logs().ERROR().Error(err)
				ctx.Render(problems.Status(ProblemNotAcceptable), ctx.Problem(ProblemNotAcceptable), true)
				return
			}
		}

		ctx.WriteHeader(http.StatusOK)
		if _, err := ctx.Write(data); err != nil {
			ctx.Logs().ERROR().Error(err)
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
	if l%2 != 0 {
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
	if l%2 != 0 {
		panic("kv 必须偶数位")
	}

	return ResponserFunc(func(ctx *Context) {
		for i := 0; i < l; i += 2 {
			ctx.Header().Add(kv[i], kv[i+1])
		}
		ctx.Render(status, body, false)
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
