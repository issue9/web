// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"net/http"

	"github.com/issue9/logs/v3"
)

const exited status = 0

type (
	// HandlerFunc 路由项处理函数原型
	//
	// 如果返回 nil，表示未出现任何错误，可以继续后续操作，
	// 非 nil，表示需要中断执行并向用户输出返回的对象。
	HandlerFunc func(*Context) Responser

	// Responser 表示向客户端输出的对象
	Responser interface {
		// Status 状态码
		Status() int

		// Headers 输出的报头
		Headers() map[string]string

		// Body 输出到 body 部分的对象
		//
		// 该对象最终经由 serialization.MarshalFunc 转换成文本输出。
		Body() interface{}
	}

	status int

	object struct {
		status  int
		headers map[string]string
		body    interface{}
	}
)

// FileServer 返回以当前模块作为文件系统的静态文件服务
func (m *Module) FileServer(name, index string) HandlerFunc {
	return m.Server().FileServer(m, name, index)
}

func (ctx *Context) renderResponser(resp Responser) {
	if resp == nil || resp == exited {
		return
	}

	if err := ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()); err != nil {
		ctx.Server().Logs().Error(err)
	}
}

func (s status) Status() int { return int(s) }

func (s status) Headers() map[string]string { return nil }

func (s status) Body() interface{} { return nil }

func (o *object) Status() int { return o.status }

func (o *object) Headers() map[string]string { return o.headers }

func (o *object) Body() interface{} { return o.body }

// Error 输出日志到 ERROR 通道并向用户输出指定状态码的页面
//
// NOTE:应该在出错的地方直接调用 Error，而不是将 Error 嵌套在另外的函数里，
// 否则出错信息的位置信息将不准确。
func (ctx *Context) Error(status int, v ...interface{}) Responser {
	ctx.Log(logs.LevelError, 2, v...)
	return Status(status)
}

// Errorf 输出日志到 ERROR 通道并向用户输出指定状态码的页面
func (ctx *Context) Errorf(status int, format string, v ...interface{}) Responser {
	ctx.Logf(logs.LevelError, 2, format, v...)
	return Status(status)
}

// Critical 输出日志到 CRITICAL 通道并向用户输出指定状态码的页面
func (ctx *Context) Critical(status int, v ...interface{}) Responser {
	ctx.Log(logs.LevelCritical, 2, v...)
	return Status(status)
}

// Criticalf 输出日志到 CRITICAL 通道并向用户输出指定状态码的页面
func (ctx *Context) Criticalf(status int, format string, v ...interface{}) Responser {
	ctx.Logf(logs.LevelCritical, 2, format, v...)
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
func Status(statusCode int) Responser {
	if statusCode < 100 || statusCode >= 600 {
		panic(fmt.Sprintf("无效的状态码 %d", statusCode))
	}
	return status(statusCode)
}

// Exit 不再执行后续操作退出当前请求
//
// 与其它返回的区别在于，Exit 表示已经向客户端输出相关内容，仅作退出。
func Exit() Responser { return exited }

// Result 返回 Result 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (ctx *Context) Result(code string, fields ResultFields) Responser {
	rslt := ctx.Server().Result(ctx.LocalePrinter, code, fields)
	return Object(rslt.Status(), rslt, nil)
}

// Redirect 重定向至新的 URL
func (ctx *Context) Redirect(status int, url string) Responser {
	http.Redirect(ctx.Response, ctx.Request, url, status)
	return Exit()
}
