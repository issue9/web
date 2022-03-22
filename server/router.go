// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"io/fs"
	"net/http"

	"github.com/issue9/logs/v3"
	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/muxutil"
)

type (
	Router         = mux.RouterOf[HandlerFunc]
	Routers        = mux.RoutersOf[HandlerFunc]
	RouterOptions  = mux.OptionsOf[HandlerFunc]
	MiddlewareFunc = mux.MiddlewareFuncOf[HandlerFunc]
	Middleware     = mux.MiddlewareOf[HandlerFunc]

	// HandlerFunc 路由项处理函数原型
	//
	// 最终调用 Responser.Apply 向客户端输出信息。
	HandlerFunc func(*Context) Responser

	// Responser 表示向客户端输出对象最终需要实现的接口
	Responser interface {
		// Apply 通过 *Context 将当前内容渲染到客户端
		Apply(*Context)
	}

	status int
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
	return ctx.Server().Result(ctx.LocalePrinter(), code, fields)
}
