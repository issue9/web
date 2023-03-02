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

	"github.com/issue9/web/internal/problems"
)

type (
	Router         = mux.RouterOf[HandlerFunc]
	Routers        = group.GroupOf[HandlerFunc]
	MiddlewareFunc = types.MiddlewareFuncOf[HandlerFunc]
	Middleware     = types.MiddlewareOf[HandlerFunc]

	// HandlerFunc 路由的处理函数
	//
	// 向客户端输出内容的有两种方法，一种是通过 [Context.Write] 方法；
	// 或是返回 [Responser] 对象。前者在调用 [Context.Write] 时即输出内容，
	// 后者会在整个请求退出时才将 [Responser] 进行编码输出。
	//
	// 返回值可以为空，表示在中间件执行过程中已经向客户端输出同内容。
	HandlerFunc func(*Context) Responser
)

func notFound(ctx *Context) Responser { return ctx.NotFound() }

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
// name 表示地址中表示文件名部分的参数名称；
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
			return ctx.Problem(problems.ProblemForbidden)
		case errors.Is(err, fs.ErrNotExist):
			return ctx.NotFound()
		case err != nil:
			srv.Logs().ERROR().Error(err)
			return ctx.Problem(problems.ProblemInternalServerError)
		default:
			return nil
		}
	}
}
