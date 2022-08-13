// SPDX-License-Identifier: MIT

package server

import (
	"net/http"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
)

type (
	// Responser 表示向客户端输出对象最终需要实现的接口
	Responser interface {
		// Apply 通过 [Context] 将当前内容渲染到客户端
		//
		// 在调用 Apply 之后，就不再使用 Responser 对象，
		// 如果你的对象支持 sync.Pool 复用，可以在 Apply 退出之际回收。
		Apply(*Context)
	}

	ResponserFunc func(*Context)
)

func (f ResponserFunc) Apply(c *Context) { f(c) }

func Status(status int, kv ...string) Responser {
	l := len(kv)
	if l%2 != 0 {
		panic("kv 必须偶数位")
	}

	return ResponserFunc(func(ctx *Context) {
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

	return ResponserFunc(func(ctx *Context) {
		for i := 0; i < l; i += 2 {
			ctx.Header().Add(kv[i], kv[i+1])
		}
		if err := ctx.Marshal(status, body, false); err != nil {
			ctx.Logs().ERROR().Error(err)
		}
	})
}

// InternalServerError 输出日志到 ERROR 通道并向用户输出 500 状态码的页面
func (ctx *Context) InternalServerError(err error) Responser {
	return ctx.err(3, http.StatusInternalServerError, err)
}

// Error 输出日志到 ERROR 通道并向用户输出指定状态码的页面
func (ctx *Context) Error(status int, err error) Responser { return ctx.err(3, status, err) }

func (ctx *Context) err(depth, status int, err error) Responser {
	entry := ctx.Logs().NewEntry(logs.LevelError).Location(depth)
	if le, ok := err.(localeutil.LocaleStringer); ok {
		entry.Message = le.LocaleString(ctx.Server().LocalePrinter())
	} else {
		entry.Message = err.Error()
	}
	ctx.Logs().Output(entry)
	return Status(status)
}
