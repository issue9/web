// SPDX-License-Identifier: MIT

package server

import (
	"fmt"

	"github.com/issue9/web/content"
)

type (
	// HandlerFunc 路由项处理函数原型
	//
	// 如果返回非空对象，则表示最终向终端输出此内容，不再需要处理其它情况。
	HandlerFunc func(*Context) Responser

	// Responser 表示向客户端输出的对象
	Responser interface {
		// Status 状态码
		Status() int

		// Headers 输出的报头
		Headers() map[string]string

		// Body 输出到 body 部分的对象
		//
		// 该对象最终经由 content.Marshal 转换成文本输出。
		Body() interface{}
	}

	status int

	object struct {
		status  int
		headers map[string]string
		body    interface{}
	}
)

func (ctx *Context) renderResponser(resp Responser) {
	if resp == nil {
		return
	}

	if err := ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()); err != nil {
		ctx.Server().Logs().Error(err)
	}
}

func (s status) Status() int { return int(s) }

func (s status) Headers() map[string]string { return nil }

func (s status) Body() interface{} { return nil }

func (o *object) Status() int { return int(o.status) }

func (o *object) Headers() map[string]string { return o.headers }

func (o *object) Body() interface{} { return o.body }

// Error 输出日志到 ERROR 通道并向用户输出指定状态码的页面
func (ctx *Context) Error(status int, v ...interface{}) Responser {
	if len(v) > 0 {
		ctx.server.Logs().ERROR().Output(2, fmt.Sprint(v...))
	}
	return Status(status)
}

// Errorf 输出日志到 ERROR 通道并向用户输出指定状态码的页面
func (ctx *Context) Errorf(status int, format string, v ...interface{}) Responser {
	if len(v) > 0 {
		ctx.server.Logs().ERROR().Output(2, fmt.Sprintf(format, v...))
	}
	return Status(status)
}

// Critical 输出日志到 CRITICAL 通道并向用户输出指定状态码的页面
func (ctx *Context) Critical(status int, v ...interface{}) Responser {
	if len(v) > 0 {
		ctx.server.Logs().CRITICAL().Output(2, fmt.Sprint(v...))
	}
	return Status(status)
}

// Criticalf 输出日志到 CRITICAL 通道并向用户输出指定状态码的页面
func (ctx *Context) Criticalf(status int, format string, v ...interface{}) Responser {
	if len(v) > 0 {
		ctx.server.Logs().CRITICAL().Output(2, fmt.Sprintf(format, v...))
	}
	return Status(status)
}

func Object(status int, body interface{}, headers map[string]string) Responser {
	return &object{
		status:  status,
		headers: headers,
		body:    body,
	}
}

// Status 仅包含状态码的 Responser
func Status(statusCode int) Responser { return status(statusCode) }

// Result 返回 Result 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (ctx *Context) Result(code int, fields content.ResultFields) Responser {
	rslt := ctx.server.Result(ctx.LocalePrinter, code, fields)
	return Object(rslt.Status(), rslt, nil)
}
