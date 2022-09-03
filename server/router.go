// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/issue9/mux/v7"
	"github.com/issue9/mux/v7/group"
	"github.com/issue9/mux/v7/types"
)

type (
	Router         = mux.RouterOf[HandlerFunc]
	Routers        = group.GroupOf[HandlerFunc]
	MiddlewareFunc = types.MiddlewareFuncOf[HandlerFunc]
	Middleware     = types.MiddlewareOf[HandlerFunc]

	// HandlerFunc 路由项处理函数原型
	HandlerFunc func(*Context) Responser

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

func notFound(ctx *Context) Responser { return ctx.Problem("404") }

func buildNodeHandle(status int) types.BuildNodeHandleOf[HandlerFunc] {
	return func(n types.Node) HandlerFunc {
		return func(ctx *Context) Responser {
			ctx.Header().Set("Allow", n.AllowHeader())
			if ctx.Request().Method == http.MethodOptions { // OPTIONS 200
				return ResponserFunc(func(ctx *Context) {
					ctx.WriteHeader(http.StatusOK)
				})
			}
			return ctx.Problem(strconv.Itoa(status))
		}
	}
}

func (srv *Server) call(w http.ResponseWriter, r *http.Request, ps types.Route, f HandlerFunc) {
	if ctx := srv.newContext(w, r, ps); ctx != nil {
		if resp := f(ctx); resp != nil {
			resp.Apply(ctx)
		}
		ctx.destroy()
	}
}

// Routers 路由管理接口
func (srv *Server) Routers() *Routers { return srv.routers }

// FileServer 构建静态文件服务对象
//
// fsys 为文件系统，如果为空则采用 [Server] 本身；
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
		p, _ := ctx.route.Params().Get(name) // 空值也是允许的值

		err := mux.ServeFile(fsys, p, index, ctx, ctx.Request())
		switch {
		case errors.Is(err, fs.ErrPermission):
			srv.Logs().WARN().Error(err)
			return ctx.Problem("403") // http.StatusForbidden
		case errors.Is(err, fs.ErrNotExist):
			srv.Logs().WARN().Error(err)
			return ctx.Problem("404") // http.StatusNotFound
		case err != nil:
			srv.Logs().ERROR().Error(err)
			return ctx.Problem("500") // http.StatusInternalServerError
		default:
			return nil
		}
	}
}
