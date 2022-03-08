// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"sync"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v3"
	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/muxutil"
)

var respPool = sync.Pool{New: func() interface{} { return &Response{} }}

type (
	// HandlerFunc 路由项处理函数原型
	//
	// 如果返回 nil，表示未出现任何错误，可以继续后续操作，
	// 非 nil，表示需要中断执行并向用户输出返回的对象。
	HandlerFunc func(*Context) *Response

	// Response 表示向客户端输出的对象
	Response struct {
		status  int
		body    any
		headers map[string]string
	}

	Router         = mux.RouterOf[HandlerFunc]
	Routers        = mux.RoutersOf[HandlerFunc]
	RouterOptions  = mux.OptionsOf[HandlerFunc]
	MiddlewareFunc = mux.MiddlewareFuncOf[HandlerFunc]
	Middleware     = mux.MiddlewareOf[HandlerFunc]
)

// Routers 路由管理接口
func (srv *Server) Routers() *Routers { return srv.routers }

// FileServer 提供静态文件服务
//
// fsys 为文件系统，如果为空则采用 srv.FS；
// name 表示参数名称；
// index 表示目录下的默认文件名；
func (srv *Server) FileServer(fsys fs.FS, name, index string) HandlerFunc {
	if fsys == nil {
		fsys = srv
	}

	if name == "" {
		panic("参数 name 不能为空")
	}

	return func(ctx *Context) *Response {
		p, _ := ctx.params.Get(name) // 空值也是允许的值

		err := muxutil.ServeFile(fsys, p, index, ctx, ctx.Request())
		switch {
		case errors.Is(err, fs.ErrPermission):
			return Resp(http.StatusForbidden)
		case errors.Is(err, fs.ErrNotExist):
			return Resp(http.StatusNotFound)
		case err != nil:
			return ctx.Error(http.StatusInternalServerError, err)
		default:
			return nil
		}
	}
}

// FileServer 返回以当前模块作为文件系统的静态文件服务
func (m *Module) FileServer(name, index string) HandlerFunc {
	return m.Server().FileServer(m, name, index)
}

func (ctx *Context) Render(resp *Response) {
	if resp == nil {
		return
	}

	if err := ctx.marshal(resp); err != nil {
		var msg string
		if ls, ok := err.(localeutil.LocaleStringer); ok {
			msg = ls.LocaleString(ctx.Server().LocalePrinter())
		} else {
			msg = err.Error()
		}
		ctx.Server().Logs().Error(msg)
	}

	respPool.Put(resp)
}

func (o *Response) Status() int { return o.status }

func (o *Response) Body() any { return o.body }

func (o *Response) SetStatus(status int) *Response {
	if status < 100 || status >= 600 {
		panic(fmt.Sprintf("无效的状态码 %d", status))
	}
	o.status = status
	return o
}

// SetBody 指定输出的对象
//
// 若是一个 nil 值，则不会向客户端输出任何内容；
// 若是需要正常输出一个 nil 类型到客户端（比如JSON 中的 null），
// 可以传递一个 *struct{} 值，或是自定义实现相应的解码函数；
func (o *Response) SetBody(body any) *Response {
	o.body = body
	return o
}

// SetHeader 设置报头
func (o *Response) SetHeader(k, v string) *Response {
	if o.headers == nil {
		o.headers = map[string]string{}
	}
	o.headers[k] = v
	return o
}

func (o *Response) GetHeader(k string) (v string, found bool) {
	v, found = o.headers[k]
	return
}

func (o *Response) DelHeader(k string) *Response {
	delete(o.headers, k)
	return o
}

// Error 输出日志到 ERROR 通道并向用户输出指定状态码的页面
//
// NOTE:应该在出错的地方直接调用 Error，而不是将 Error 嵌套在另外的函数里，
// 否则出错信息的位置信息将不准确。
func (ctx *Context) Error(status int, v ...any) *Response {
	ctx.Log(logs.LevelError, 2, v...)
	return Resp(status)
}

// Errorf 输出日志到 ERROR 通道并向用户输出指定状态码的页面
func (ctx *Context) Errorf(status int, format string, v ...any) *Response {
	ctx.Logf(logs.LevelError, 2, format, v...)
	return Resp(status)
}

// Critical 输出日志到 CRITICAL 通道并向用户输出指定状态码的页面
func (ctx *Context) Critical(status int, v ...any) *Response {
	ctx.Log(logs.LevelCritical, 2, v...)
	return Resp(status)
}

// Criticalf 输出日志到 CRITICAL 通道并向用户输出指定状态码的页面
func (ctx *Context) Criticalf(status int, format string, v ...any) *Response {
	ctx.Logf(logs.LevelCritical, 2, format, v...)
	return Resp(status)
}

func Resp(status int) *Response {
	resp := respPool.Get().(*Response)
	resp.body = nil
	resp.headers = nil
	resp.status = 0
	return resp.SetStatus(status)
}

// Result 向客户端输出指定代码的错误信息
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (ctx *Context) Result(code string, fields ResultFields) *Response {
	rslt := ctx.Server().Result(ctx.LocalePrinter, code, fields)
	return Resp(rslt.Status()).SetBody(rslt)
}

// Redirect 重定向至新的 URL
func (ctx *Context) Redirect(status int, url string) *Response {
	http.Redirect(ctx, ctx.Request(), url, status)
	return nil
}
