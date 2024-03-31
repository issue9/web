// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v8"
	"github.com/issue9/mux/v8/group"
	"github.com/issue9/mux/v8/header"
	"github.com/issue9/mux/v8/types"
	"github.com/issue9/source"

	"github.com/issue9/web/internal/errs"
)

type (
	Router            = mux.RouterOf[HandlerFunc]
	Prefix            = mux.PrefixOf[HandlerFunc]
	Resource          = mux.ResourceOf[HandlerFunc]
	RouterMatcher     = group.Matcher
	RouterMatcherFunc = group.MatcherFunc
	RouterOption      = mux.Option
	MiddlewareFunc    = types.MiddlewareFuncOf[HandlerFunc]
	Middleware        = types.MiddlewareOf[HandlerFunc]

	// HandlerFunc 路由的处理函数原型
	//
	// 向客户端输出内容的有两种方法，一种是通过 [Context.Write] 方法；
	// 或是返回 [Responser] 对象。前者在调用 [Context.Write] 时即输出内容，
	// 后者会在整个请求退出时才将 [Responser] 进行编码输出。
	//
	// 返回值可以为空，表示在中间件执行过程中已经向客户端输出同内容。
	HandlerFunc = func(*Context) Responser

	// Routers 提供管理路由的接口
	Routers struct {
		g *group.GroupOf[HandlerFunc]
	}
)

func notFound(ctx *Context) Responser { return ctx.NotFound() }

func buildNodeHandle(status int) types.BuildNodeHandleOf[HandlerFunc] {
	return func(n types.Node) HandlerFunc {
		return func(ctx *Context) Responser {
			ctx.Header().Set(header.Allow, n.AllowHeader())
			if ctx.Request().Method == http.MethodOptions { // OPTIONS 200
				return ResponserFunc(func(ctx *Context) {
					ctx.WriteHeader(http.StatusOK)
				})
			}
			return ctx.Problem(strconv.Itoa(status))
		}
	}
}

func (s *InternalServer) call(w http.ResponseWriter, r *http.Request, route types.Route, f HandlerFunc) {
	if ctx := s.NewContext(w, r, route); ctx != nil {
		if resp := f(ctx); resp != nil {
			resp.Apply(ctx)
		}
		s.freeContext(ctx)
	}
}

func (s *InternalServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.routers.g.ServeHTTP(w, r)
}

func (s *InternalServer) Routers() *Routers { return s.routers }

// Get 获取指定名称的路由
//
// name 指由 [Routers.New] 的 name 参数指定的值；
func (r *Routers) Get(name string) *Router { return r.g.Router(name) }

// New 声明新路由
func (r *Routers) New(name string, matcher RouterMatcher, o ...RouterOption) *Router {
	return r.g.New(name, matcher, o...)
}

// Remove 删除指定名称的路由
//
// name 指由 [Routers.New] 的 name 参数指定的值；
func (r *Routers) Remove(name string) { r.g.Remove(name) }

// Routers 返回所有的路由
func (r *Routers) Routers() []*Router { return r.g.Routers() }

// Use 对所有的路由使用中间件
func (r *Routers) Use(m ...Middleware) { r.g.Use(m...) }

// Recovery 在路由奔溃之后的处理方式
//
// 相对于 [mux.Recovery] 的相关功能，提供了对 [web.NewError] 错误的处理。
func Recovery(status int, l *logs.Logger) RouterOption {
	return mux.Recovery(func(w http.ResponseWriter, msg any) {
		err, ok := msg.(error)
		if !ok {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			l.String(source.Stack(4, err))
			return
		}

		he := &errs.HTTP{}
		if !errors.As(err, &he) {
			he.Status = status
			he.Message = err
		}
		http.Error(w, http.StatusText(he.Status), he.Status)
		l.String(source.Stack(4, he.Message))
	})
}

// CORS 自定义跨域请求设置项
//
// 具体参数可参考 [mux.CORS]。
func CORS(origin, allowHeaders, exposedHeaders []string, maxAge int, allowCredentials bool) RouterOption {
	return mux.CORS(origin, allowHeaders, exposedHeaders, maxAge, allowCredentials)
}

// DenyCORS 禁用跨域请求
func DenyCORS() RouterOption { return mux.DenyCORS() }

// AllowedCORS 允许跨域请求
func AllowedCORS(maxAge int) RouterOption { return mux.AllowedCORS(maxAge) }

// URLDomain 为 [RouterOf.URL] 生成的地址带上域名
func URLDomain(prefix string) RouterOption { return mux.URLDomain(prefix) }
