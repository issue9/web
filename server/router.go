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

// FileServer 构建静态文件服务对象
//
// fsys 为文件系统，如果为空则采用 [Server] 本身；
// name 表示参数名称；
// index 表示目录下的默认文件名；
// problems 表示指定状态需要关联的错误 id。默认情况下，
// 文件服务采用 [Status] 返回指定的状态码。如果需要将返回内容转换成 [Problem] 对象，
// 可以在此参数中指定，其中键名为状态码，键值为对应的错误 id，可以用的状态有：
//   - http.StatusForbidden
//   - http.StatusNotFound
//   - http.StatusInternalServerError
func (srv *Server) FileServer(fsys fs.FS, name, index string, problems map[int]string) HandlerFunc {
	if fsys == nil {
		fsys = srv
	}

	if problems == nil {
		problems = map[int]string{}
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
			if id := problems[http.StatusForbidden]; id != "" {
				return ctx.Server().Problems().Problem(id)
			}
			return Status(http.StatusForbidden)
		case errors.Is(err, fs.ErrNotExist):
			srv.Logs().WARN().Error(err)
			if id := problems[http.StatusNotFound]; id != "" {
				return ctx.Server().Problems().Problem(id)
			}
			return Status(http.StatusNotFound)
		case err != nil:
			srv.Logs().ERROR().Error(err)
			if id := problems[http.StatusInternalServerError]; id != "" {
				return ctx.Server().Problems().Problem(id)
			}
			return Status(http.StatusInternalServerError)
		default:
			return nil
		}
	}
}
