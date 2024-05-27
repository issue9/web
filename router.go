// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/issue9/mux/v9"
	"github.com/issue9/mux/v9/header"
	"github.com/issue9/mux/v9/types"
	"github.com/issue9/source"

	"github.com/issue9/web/internal/errs"
)

type (
	Router            = mux.Router[HandlerFunc]
	Prefix            = mux.Prefix[HandlerFunc]
	Resource          = mux.Resource[HandlerFunc]
	RouterMatcher     = mux.Matcher
	RouterMatcherFunc = mux.MatcherFunc
	RouterOption      = mux.Option
	MiddlewareFunc    = types.MiddlewareFunc[HandlerFunc]
	Middleware        = types.Middleware[HandlerFunc]

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
		g *mux.Group[HandlerFunc]
	}
)

func notFound(ctx *Context) Responser { return ctx.NotFound() }

func buildNodeHandle(status int) types.BuildNodeHandler[HandlerFunc] {
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

// WithRecovery 在路由奔溃之后的处理方式
//
// 相对于 [mux.WithRecovery]，提供了对 [NewError] 错误的处理。
func WithRecovery(status int, l *Logger) RouterOption {
	return mux.WithRecovery(func(w http.ResponseWriter, msg any) {
		err, ok := msg.(error)
		if !ok {
			http.Error(w, http.StatusText(status), status)
			l.String(source.Stack(4, true, err))
			return
		}

		he := &errs.HTTP{}
		if !errors.As(err, &he) {
			he.Status = status
			he.Message = err
		}
		http.Error(w, http.StatusText(he.Status), he.Status)
		l.String(source.Stack(4, true, he.Message))
	})
}

// WithCORS 自定义跨域请求设置项
//
// 具体参数可参考 [mux.WithCORS]。
func WithCORS(origin, allowHeaders, exposedHeaders []string, maxAge int, allowCredentials bool) RouterOption {
	return mux.WithCORS(origin, allowHeaders, exposedHeaders, maxAge, allowCredentials)
}

// WithDenyCORS 禁用跨域请求
func WithDenyCORS() RouterOption { return mux.WithDenyCORS() }

// WithAllowedCORS 允许跨域请求
func WithAllowedCORS(maxAge int) RouterOption { return mux.WithAllowedCORS(maxAge) }

// WithURLDomain 为 [Router.URL] 生成的地址带上域名
func WithURLDomain(prefix string) RouterOption { return mux.WithURLDomain(prefix) }

// WithTrace 控制 TRACE 请求是否有效
//
// body 表示是否显示 body 内容；
func WithTrace(body bool) RouterOption {
	return mux.WithTrace(func(ctx *Context) Responser {
		mux.Trace(ctx, ctx.Request(), body)
		return nil
	})
}

func WithAnyInterceptor(rule string) RouterOption { return WithAnyInterceptor(rule) }

func WithDigitInterceptor(rule string) RouterOption { return mux.WithDigitInterceptor(rule) }

func WithWordInterceptor(rule string) RouterOption { return mux.WithWordInterceptor(rule) }

func WithInterceptor(f mux.InterceptorFunc, rule ...string) RouterOption {
	return mux.WithInterceptor(f, rule...)
}
