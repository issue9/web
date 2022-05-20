// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"io/fs"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/muxutil"
)

var objectPool = &sync.Pool{New: func() any { return &object{} }}

type (
	Router         = mux.RouterOf[HandlerFunc]
	Routers        = mux.RoutersOf[HandlerFunc]
	RouterOptions  = mux.Options
	MiddlewareFunc = mux.MiddlewareFuncOf[HandlerFunc]
	Middleware     = mux.MiddlewareOf[HandlerFunc]

	// HandlerFunc 路由项处理函数原型
	//
	// 最终调用 Responser.Apply 向客户端输出信息。
	HandlerFunc func(*Context) Responser

	// Responser 表示向客户端输出对象最终需要实现的接口
	Responser interface {
		// Apply 通过 *Context 将当前内容渲染到客户端
		//
		// 在调用 Apply 之后，就不再使用 Responser 对象，
		// 如果你的对象支持 sync.Pool 复用，可以在 Apply 退出之际进行回收至 sync.Pool。
		Apply(*Context)
	}

	status int

	object struct {
		status int
		body   any
	}
)

// Status 仅向客户端输出状态码
func Status(code int) Responser { return status(code) }

func (s status) Apply(ctx *Context) { ctx.WriteHeader(int(s)) }

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

	return func(ctx *Context) Responser {
		p, _ := ctx.params.Get(name) // 空值也是允许的值

		err := muxutil.ServeFile(fsys, p, index, ctx, ctx.Request())
		switch {
		case errors.Is(err, fs.ErrPermission):
			srv.Logs().WARN().Error(err)
			return Status(http.StatusForbidden)
		case errors.Is(err, fs.ErrNotExist):
			srv.Logs().WARN().Error(err)
			return Status(http.StatusNotFound)
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

// Result 向客户端输出指定代码的错误信息
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (ctx *Context) Result(code string, fields ResultFields) Responser {
	return ctx.Server().Result(ctx.LocalePrinter(), code, fields)
}

// Status 仅向客户端输出状态码和报头
//
// kv 为报头，必须以偶数数量出现，奇数位为报头名，偶数位为对应的报头值；
func (ctx *Context) Status(code int, kv ...string) Responser {
	if l := len(kv); l > 0 {
		if l%2 != 0 {
			panic("kv 必须偶数位")
		}
		for i := 0; i < l; i += 2 {
			ctx.Header().Add(kv[i], kv[i+1])
		}
	}
	return status(code)
}

// Object 输出状态和对象至客户端
//
// body 表示需要输出的对象，该对象最终会被转换成相应的编码；
// kv 为报头，必须以偶数数量出现，奇数位为报头名，偶数位为对应的报头值；
func (ctx *Context) Object(status int, body interface{}, kv ...string) Responser {
	o := objectPool.Get().(*object)
	o.status = status
	o.body = body

	if l := len(kv); l > 0 {
		if l%2 != 0 {
			panic("kv 必须偶数位")
		}
		for i := 0; i < l; i += 2 {
			ctx.Header().Add(kv[i], kv[i+1])
		}
	}
	return o
}

func (o *object) Apply(ctx *Context) {
	if err := ctx.Marshal(o.status, o.body); err != nil {
		ctx.Logs().ERROR().Error(err)
	}
	objectPool.Put(o)
}

func (ctx *Context) Created(v any, location string) Responser {
	if location != "" {
		return ctx.Object(http.StatusCreated, v, "Location", location)
	}
	return ctx.Object(http.StatusCreated, v)
}

// OK 返回 200 状态码下的对象
func (ctx *Context) OK(v any) Responser { return ctx.Object(http.StatusOK, v) }

func (ctx *Context) NotFound() Responser { return Status(http.StatusNotFound) }

func (ctx *Context) NoContent() Responser { return Status(http.StatusNoContent) }

func (ctx *Context) NotImplemented() Responser { return Status(http.StatusNotImplemented) }

// RetryAfter 返回 Retry-After 报头内容
//
// 一般适用于 301 和 503 报文。
//
// status 表示返回的状态码；seconds 表示秒数，如果想定义为时间格式，
// 可以采用 RetryAt 函数，两个功能是相同的，仅是时间格式上有差别。
func (ctx *Context) RetryAfter(status int, seconds uint64) Responser {
	return ctx.Status(status, "Retry-After", strconv.FormatUint(seconds, 10))
}

func (ctx *Context) RetryAt(status int, at time.Time) Responser {
	return ctx.Status(status, "Retry-After", at.UTC().Format(http.TimeFormat))
}

// Redirect 重定向至新的 URL
func (ctx *Context) Redirect(status int, url string) Responser {
	return ctx.Status(status, "Location", url)
}
