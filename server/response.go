// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"net/http"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v3"
)

type (
	// HandlerFunc 路由项处理函数原型
	//
	// 如果返回 nil，表示未出现任何错误，可以继续后续操作，
	// 非 nil，表示需要中断执行并向用户输出返回的对象。
	HandlerFunc func(*Context) *Responser

	// Responser 表示向客户端输出的对象
	Responser struct {
		status  int
		body    any
		headers map[string]string
	}
)

// FileServer 返回以当前模块作为文件系统的静态文件服务
func (m *Module) FileServer(name, index string) HandlerFunc {
	return m.Server().FileServer(m, name, index)
}

func (ctx *Context) Render(resp *Responser) {
	if resp == nil {
		return
	}

	if err := ctx.Marshal(resp.status, resp.body, resp.headers); err != nil {
		var msg string
		if ls, ok := err.(localeutil.LocaleStringer); ok {
			msg = ls.LocaleString(ctx.Server().LocalePrinter())
		} else {
			msg = err.Error()
		}
		ctx.Server().Logs().Error(msg)
	}
}

func (o *Responser) Status(status int) *Responser {
	o.status = status
	return o
}

func (o *Responser) Body(body any) *Responser {
	o.body = body
	return o
}

func (o *Responser) SetHeader(k, v string) *Responser {
	if o.headers == nil {
		o.headers = map[string]string{}
	}
	o.headers[k] = v
	return o
}

func (o *Responser) GetHeader(k string) string {
	return o.headers[k]
}

// Error 输出日志到 ERROR 通道并向用户输出指定状态码的页面
//
// NOTE:应该在出错的地方直接调用 Error，而不是将 Error 嵌套在另外的函数里，
// 否则出错信息的位置信息将不准确。
func (ctx *Context) Error(status int, v ...any) *Responser {
	ctx.Log(logs.LevelError, 2, v...)
	return Status(status)
}

// Errorf 输出日志到 ERROR 通道并向用户输出指定状态码的页面
func (ctx *Context) Errorf(status int, format string, v ...any) *Responser {
	ctx.Logf(logs.LevelError, 2, format, v...)
	return Status(status)
}

// Critical 输出日志到 CRITICAL 通道并向用户输出指定状态码的页面
func (ctx *Context) Critical(status int, v ...any) *Responser {
	ctx.Log(logs.LevelCritical, 2, v...)
	return Status(status)
}

// Criticalf 输出日志到 CRITICAL 通道并向用户输出指定状态码的页面
func (ctx *Context) Criticalf(status int, format string, v ...any) *Responser {
	ctx.Logf(logs.LevelCritical, 2, format, v...)
	return Status(status)
}

func Status(status int) *Responser {
	if status < 100 || status >= 600 {
		panic(fmt.Sprintf("无效的状态码 %d", status))
	}
	return &Responser{status: status}
}

// Result 返回 Result 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (ctx *Context) Result(code string, fields ResultFields) *Responser {
	rslt := ctx.Server().Result(ctx.LocalePrinter, code, fields)
	return Status(rslt.Status()).Body(rslt)
}

// Redirect 重定向至新的 URL
func (ctx *Context) Redirect(status int, url string) *Responser {
	http.Redirect(ctx, ctx.Request(), url, status)
	return nil
}
