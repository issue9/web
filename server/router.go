// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v7"
	"github.com/issue9/mux/v7/types"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/internal/header"
)

func notFound(ctx *web.Context) web.Responser { return ctx.NotFound() }

func buildNodeHandle(status int) types.BuildNodeHandleOf[web.HandlerFunc] {
	return func(n types.Node) web.HandlerFunc {
		return func(ctx *web.Context) web.Responser {
			ctx.Header().Set(header.Allow, n.AllowHeader())
			if ctx.Request().Method == http.MethodOptions { // OPTIONS 200
				return web.ResponserFunc(func(ctx *web.Context) {
					ctx.WriteHeader(http.StatusOK)
				})
			}
			return ctx.Problem(strconv.Itoa(status))
		}
	}
}

func (srv *httpServer) call(w http.ResponseWriter, r *http.Request, route types.Route, f web.HandlerFunc) {
	if ctx := srv.NewContext(w, r, route); ctx != nil {
		if resp := f(ctx); resp != nil {
			resp.Apply(ctx)
		}
		srv.s.FreeContext(ctx)
	}
}

func (srv *httpServer) NewContext(w http.ResponseWriter, r *http.Request, route types.Route) *web.Context {
	return srv.s.NewContext(w, r, route)
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

// Recovery 在路由奔溃之后的处理方式
//
// 相对于 [mux.Recovery] 的相关功能，提供了对 [web.NewError] 错误的处理。
func Recovery(l *logs.Logger) mux.Option {
	return mux.Recovery(func(w http.ResponseWriter, msg any) {
		err, ok := msg.(error)
		if !ok {
			l.Print(msg)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		he := &errs.HTTP{}
		if !errors.As(err, &he) {
			he.Status = http.StatusInternalServerError
			he.Message = err
		}
		l.Error(he.Message)
		w.WriteHeader(he.Status)
	})
}
