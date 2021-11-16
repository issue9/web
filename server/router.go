// SPDX-License-Identifier: MIT

package server

import (
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/issue9/middleware/v5/debugger"
	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/group"
)

type (
	// Router 路由
	Router struct {
		srv      *Server
		router   *mux.Router
		filters  []Filter
		debugger *debugger.Debugger
	}

	// Prefix 带有统一前缀的路由管理
	Prefix struct {
		router  *Router
		prefix  string
		filters []Filter
	}
)

// NewRouter 构建基于 matcher 匹配的路由操作实例
//
// domain 仅用于 URL 生成地址，并不会对路由本身产生影响。
func (srv *Server) NewRouter(name, domain string, matcher group.Matcher, filter ...Filter) (*Router, error) {
	r := srv.MuxGroup().New(name, matcher, mux.URLDomain(domain))
	dbg := &debugger.Debugger{}
	r.Middlewares().Append(dbg.Middleware)
	rr := &Router{
		srv:      srv,
		router:   r,
		filters:  filter,
		debugger: dbg,
	}
	srv.routers[name] = rr

	return rr, nil
}

// Router 返回由 Server.NewRouter 声明的路由
func (srv *Server) Router(name string) *Router { return srv.routers[name] }

func (srv *Server) RemoveRouter(name string) {
	srv.MuxGroup().Remove(name)
	delete(srv.routers, name)
}

// MuxGroup 返回 group.Groups 实例
func (srv *Server) MuxGroup() *group.Group { return srv.group }

// SetDebugger 设置调试地址
func (router *Router) SetDebugger(pprof, vars string) {
	router.debugger.Pprof = pprof
	router.debugger.Vars = vars
}

func (router *Router) handleWithFilters(path string, h HandlerFunc, filters []Filter, method ...string) {
	h = ApplyFilters(h, filters...)
	router.router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := router.srv.NewContext(w, r)
		ctx.renderResponser(h(ctx))
		contextPool.Put(ctx)
	}, method...)
}

func (router *Router) Handle(path string, h HandlerFunc, method ...string) *Router {
	router.handleWithFilters(path, h, router.filters, method...)
	return router
}

func (router *Router) MuxRouter() *mux.Router { return router.router }

func (router *Router) Get(path string, h HandlerFunc) *Router {
	return router.Handle(path, h, http.MethodGet)
}

func (router *Router) Post(path string, h HandlerFunc) *Router {
	return router.Handle(path, h, http.MethodPost)
}

func (router *Router) Put(path string, h HandlerFunc) *Router {
	return router.Handle(path, h, http.MethodPut)
}

func (router *Router) Delete(path string, h HandlerFunc) *Router {
	return router.Handle(path, h, http.MethodDelete)
}

func (router *Router) Patch(path string, h HandlerFunc) *Router {
	return router.Handle(path, h, http.MethodPatch)
}

func (router *Router) Remove(path string, method ...string) {
	router.router.Remove(path, method...)
}

func (router *Router) clone() *Router {
	filters := make([]Filter, len(router.filters))
	copy(filters, router.filters)
	return &Router{
		srv:     router.srv,
		router:  router.router,
		filters: filters,
	}
}

// URL 构建完整的 URL
//
// 功能与 mux.Router.URL 相似，但是加上了关联的域名地址。比如根地址是 https://example.com/blog
// pattern 为 /posts/{id}，则返回为 https://example.com/blog/posts/1。
// 如果 params 为空的话，则会直接将 pattern 作为从 mux 转换之后的内容与 router.root 合并返回。
func (router *Router) URL(strict bool, pattern string, params map[string]string) (string, error) {
	return router.router.URL(strict, pattern, params)
}

// Static 添加静态路由
//
// p 为路由地址，必须以命名参数结尾，比如 /assets/{path}，之后可以通过此值删除路由项；
// dir 为指向静态文件的路径；
// index 可以在访问一个目录时指定默认访问的页面。
//
// 如果要删除该静态路由，则可以将 path 传递给 Remove 进行删除。
//
// 比如在 Root 的值为 example.com/blog 时，
// 将参数指定为 /admin/{path} 和 ~/data/assets/admin
// 表示将 example.com/blog/admin/* 解析到 ~/data/assets/admin 目录之下。
func (router *Router) Static(p, dir, index string) {
	router.StaticFS(p, os.DirFS(dir), index)
}

func (router *Router) StaticFS(p string, f fs.FS, index string) {
	lastStart := strings.LastIndexByte(p, '{')
	if lastStart < 0 || len(p) == 0 || p[len(p)-1] != '}' || lastStart+2 == len(p) {
		panic("path 必须是命名参数结尾：比如 /assets/{path}。")
	}

	router.Handle(p, func(ctx *Context) Responser {
		pp := ctx.Request.URL.Path
		pp = strings.TrimPrefix(pp, p[:lastStart])
		if pp != "" && pp[0] == '/' {
			pp = pp[1:]
		}
		return ctx.ServeFileFS(f, pp, index, nil)
	}, http.MethodGet)
}

// Prefix 返回特定前缀的路由设置对象
func (router *Router) Prefix(prefix string, filter ...Filter) *Prefix {
	return &Prefix{
		router:  router.clone(),
		prefix:  prefix,
		filters: filter,
	}
}

// Handle 添加路由项
func (p *Prefix) Handle(path string, h HandlerFunc, method ...string) *Prefix {
	p.router.handleWithFilters(p.prefix+path, h, p.filters, method...)
	return p
}

func (p *Prefix) Get(path string, h HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodGet)
}

func (p *Prefix) Post(path string, h HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodPost)
}

func (p *Prefix) Put(path string, h HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodPut)
}

func (p *Prefix) Delete(path string, h HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodDelete)
}

func (p *Prefix) Patch(path string, h HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodPatch)
}

func (p *Prefix) Remove(path string, method ...string) {
	p.router.Remove(p.prefix+path, method...)
}

func (p *Prefix) URL(strict bool, path string, params map[string]string) (string, error) {
	return p.router.URL(strict, p.prefix+path, params)
}
