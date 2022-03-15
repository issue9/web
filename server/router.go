// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"io/fs"
	"net/http"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v3"
	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/muxutil"
)

type (
	// HandlerFunc 路由项处理函数原型
	//
	// 如果返回 nil，表示未出现任何错误，可以继续后续操作，
	// 非 nil，表示需要中断执行并向用户输出返回的对象。
	HandlerFunc func(*Context) Responser

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

	return func(ctx *Context) Responser {
		p, _ := ctx.params.Get(name) // 空值也是允许的值

		err := muxutil.ServeFile(fsys, p, index, ctx, ctx.Request())
		switch {
		case errors.Is(err, fs.ErrPermission):
			return Status(http.StatusForbidden)
		case errors.Is(err, fs.ErrNotExist):
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

func (ctx *Context) Render(resp Responser) {
	if resp == nil {
		return
	}

	if err := resp.Apply(ctx); err != nil {
		var msg string
		if ls, ok := err.(localeutil.LocaleStringer); ok {
			msg = ls.LocaleString(ctx.Server().LocalePrinter())
		} else {
			msg = err.Error()
		}
		ctx.Server().Logs().Error(msg)
	}
}

// InternalServerError 输出日志到 ERROR 通道并向用户输出 500 状态码的页面
//
// 注意事项参考 Error
func (ctx *Context) InternalServerError(v ...any) Responser {
	ctx.Log(logs.LevelError, 2, v...)
	return Status(http.StatusInternalServerError)
}

// InternalServerErrorf 输出日志到 ERROR 通道并向用户输出 500 状态码的页面
//
// 注意事项参考 Error
func (ctx *Context) InternalServerErrorf(format string, v ...any) Responser {
	ctx.Logf(logs.LevelError, 2, format, v...)
	return Status(http.StatusInternalServerError)
}

// Error 输出日志到 ERROR 通道并向用户输出指定状态码的页面
//
// NOTE:应该在出错的地方直接调用 Error，而不是将 Error 嵌套在另外的函数里，
// 否则出错信息的位置信息将不准确。
func (ctx *Context) Error(status int, v ...any) Responser {
	ctx.Log(logs.LevelError, 2, v...)
	return Status(status)
}

// Errorf 输出日志到 ERROR 通道并向用户输出指定状态码的页面
func (ctx *Context) Errorf(status int, format string, v ...any) Responser {
	ctx.Logf(logs.LevelError, 2, format, v...)
	return Status(status)
}

// Result 向客户端输出指定代码的错误信息
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (ctx *Context) Result(code string, fields ResultFields) Responser {
	return ctx.Server().Result(ctx.LocalePrinter, code, fields)
}

// Redirect 重定向至新的 URL
func (ctx *Context) Redirect(status int, url string) Responser {
	return Body(status, nil).Header("Location", url)
}
