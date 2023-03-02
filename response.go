// SPDX-License-Identifier: MIT

package web

import "net/http"

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

// Object 输出状态和对象至客户端
//
// body 表示需要输出的对象，该对象最终会被转换成相应的编码；
// kv 为报头，必须以偶数数量出现，奇数位为报头名，偶数位为对应的报头值；
func Object(status int, body any, kv ...string) Responser {
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
		return Object(http.StatusCreated, v, "Location", location)
	}
	return Object(http.StatusCreated, v)
}

// OK 返回 200 状态码下的对象
func OK(v any) Responser { return Object(http.StatusOK, v) }

func NoContent() Responser { return Status(http.StatusNoContent) }

// Redirect 重定向至新的 URL
func Redirect(status int, url string) Responser {
	return Status(status, "Location", url)
}
