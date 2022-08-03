// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"io/fs"
	"net/http"

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
)

func notFound(*Context) Responser { return Status(http.StatusNotFound) }

func buildNodeHandle(status int) types.BuildNodeHandleOf[HandlerFunc] {
	return func(n types.Node) HandlerFunc {
		return func(ctx *Context) Responser {
			ctx.Header().Set("Allow", n.AllowHeader())
			return Status(status)
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
		p, _ := ctx.route.Params().Get(name) // 空值也是允许的值

		err := mux.ServeFile(fsys, p, index, ctx, ctx.Request())
		switch {
		case errors.Is(err, fs.ErrPermission):
			srv.Logs().WARN().Error(err)
			return Status(http.StatusForbidden)
		case errors.Is(err, fs.ErrNotExist):
			srv.Logs().WARN().Error(err)
			return Status(http.StatusNotFound)
		case err != nil:
			srv.Logs().ERROR().Error(err)
			return Status(http.StatusInternalServerError)
		default:
			return nil
		}
	}
}
