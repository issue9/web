// SPDX-License-Identifier: MIT

// Package response Responser 接口的前置声明
package response

import (
	"net/http"

	"github.com/issue9/logs/v4"
)

// Context 这是 [Responser] 用到的最小接口范围
type Context interface {
	http.ResponseWriter

	Marshal(int, any) error

	Logs() *logs.Logs
}

// Responser 表示向客户端输出对象最终需要实现的接口
type Responser interface {
	// Apply 通过 [Context] 将当前内容渲染到客户端
	//
	// 在调用 Apply 之后，就不再使用 [Responser] 对象，
	// 如果你的对象支持 sync.Pool 复用，可以在 Apply 退出之际进行回收至 sync.Pool。
	Apply(Context)
}

type ResponserFunc func(Context)

func (f ResponserFunc) Apply(c Context) { f(c) }

func Status(status int, kv ...string) Responser {
	l := len(kv)
	if l%2 != 0 {
		panic("kv 必须偶数位")
	}

	return ResponserFunc(func(ctx Context) {
		for i := 0; i < l; i += 2 {
			ctx.Header().Add(kv[i], kv[i+1])
		}

		ctx.WriteHeader(status)
	})
}

func Object(status int, body any, kv ...string) Responser {
	l := len(kv)
	if l%2 != 0 {
		panic("kv 必须偶数位")
	}

	return ResponserFunc(func(ctx Context) {
		for i := 0; i < l; i += 2 {
			ctx.Header().Add(kv[i], kv[i+1])
		}

		if err := ctx.Marshal(status, body); err != nil {
			ctx.Logs().ERROR().Error(err)
		}
	})
}
