// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"strconv"

	"github.com/issue9/mux/v7/types"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/header"
)

func notFound(ctx *web.Context) web.Responser { return ctx.NotFound() }

func buildNodeHandle(status int) types.BuildNodeHandleOf[web.HandlerFunc] {
	return func(n types.Node) web.HandlerFunc {
		return func(ctx *web.Context) web.Responser {
			ctx.Header().Set(header.Allow, n.AllowHeader())
			if ctx.Request().Method == http.MethodOptions { // OPTIONS 200
				return web.ResponserFunc(func(ctx *web.Context) *web.Problem {
					ctx.WriteHeader(http.StatusOK)
					return nil
				})
			}
			return ctx.Problem(strconv.Itoa(status))
		}
	}
}

func (srv *httpServer) call(w http.ResponseWriter, r *http.Request, route types.Route, f web.HandlerFunc) {
	if ctx := srv.NewContext(w, r, route); ctx != nil {
		if resp := f(ctx); resp != nil {
			if p := resp.Apply(ctx); p != nil {
				p.Apply(ctx) // Problem.Apply 始终返回 nil
			}
		}
		srv.ctxBuilder.FreeContext(ctx)
	}
}

func (srv *httpServer) NewContext(w http.ResponseWriter, r *http.Request, route types.Route) *web.Context {
	return srv.ctxBuilder.NewContext(w, r, route)
}

// NewRouter 声明新路由
//
// 新路由会继承 [Options.RoutersOptions] 中指定的参数，其中的 o 可以覆盖相同的参数；
func (srv *httpServer) NewRouter(name string, matcher web.RouterMatcher, o ...web.RouterOption) *web.Router {
	return srv.routers.New(name, matcher, o...)
}

// RemoveRouter 删除指定名称的路由
//
// name 指由 [Server.NewRouter] 的 name 参数指定的值；
func (srv *httpServer) RemoveRouter(name string) { srv.routers.Remove(name) }

// GetRouter 返回指定名称的路由
//
// name 指由 [Server.NewRouter] 的 name 参数指定的值；
func (srv *httpServer) GetRouter(name string) *web.Router { return srv.routers.Router(name) }

// UseMiddleware 为所有的路由添加新的中间件
func (srv *httpServer) UseMiddleware(m ...web.Middleware) { srv.routers.Use(m...) }

// Routers 返回路由列表
func (srv *httpServer) Routers() []*web.Router { return srv.routers.Routers() }
