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
	Prefix         = mux.PrefixOf[HandlerFunc]
	Resource       = mux.ResourceOf[HandlerFunc]
	RouterMatcher  = group.Matcher
	RouterOption   = mux.Option
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
				return ResponserFunc(func(ctx *Context) *Problem {
					ctx.WriteHeader(http.StatusOK)
					return nil
				})
			}
			return ctx.Problem(strconv.Itoa(status))
		}
	}
}

func (srv *Server) call(w http.ResponseWriter, r *http.Request, ps types.Route, f HandlerFunc) {
	if ctx := srv.newContext(w, r, ps); ctx != nil {
		if resp := f(ctx); resp != nil {
			if p := resp.Apply(ctx); p != nil {
				p.Apply(ctx) // Problem.Apply 始终返回 nil
			}
		}
		ctx.destroy()
	}
}

// NewRouter 声明新路由
//
// 新路由会继承 [Options.RoutersOptions] 中指定的参数，其中的 o 可以覆盖相同的参数；
func (srv *Server) NewRouter(name string, matcher RouterMatcher, o ...RouterOption) *Router {
	return srv.routers.New(name, matcher, o...)
}

// RemoveRouter 删除指定名称的路由
//
// name 指由 [Server.NewRouter] 的 name 参数指定的值；
func (srv *Server) RemoveRouter(name string) { srv.routers.Remove(name) }

// GetRouter 返回指定名称的路由
//
// name 指由 [Server.NewRouter] 的 name 参数指定的值；
func (srv *Server) GetRouter(name string) *Router { return srv.routers.Router(name) }

// UseMiddleware 为所有的路由添加新的中间件
func (srv *Server) UseMiddleware(m ...Middleware) { srv.routers.Use(m...) }

// Routers 返回路由列表
func (srv *Server) Routers() []*Router { return srv.routers.Routers() }

// FileServer 构建静态文件服务对象
//
// fsys 为文件系统，如果为空则采用 [Server] 本身；
// name 表示地址中表示文件名部分的参数名称；
// index 表示目录下的默认文件名；
func (srv *Server) FileServer(fsys fs.FS, name, index string) HandlerFunc {
	if name == "" {
		panic("参数 name 不能为空")
	}

	if fsys == nil {
		fsys = srv
	}

	return func(ctx *Context) Responser {
		p, _ := ctx.Route().Params().Get(name) // 空值也是允许的值

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
