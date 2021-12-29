// SPDX-License-Identifier: MIT

package server

import (
	"net/http"

	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/group"
)

type (
	// Router 路由
	Router struct {
		srv     *Server
		router  *mux.Router
		filters []Filter
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
// domain 仅用于 URL 生成地址，并不会对路由本身产生影响，可以为空。
func (srv *Server) NewRouter(name, domain string, matcher group.Matcher, filter ...Filter) *Router {
	r := srv.group.New(name, matcher, mux.URLDomain(domain))
	rr := &Router{
		srv:     srv,
		router:  r,
		filters: filter,
	}
	srv.routers[name] = rr

	return rr
}

// Routes 返回所有路由的注册路由项
//
// 第一个键名表示路由名称，第二键值表示路由项地址，值表示该路由项支持的请求方法；
func (srv *Server) Routes() map[string]map[string][]string {
	return srv.group.Routes()
}

func (srv *Server) Routers() []*Router {
	routers := make([]*Router, 0, len(srv.routers))
	for _, router := range srv.routers {
		routers = append(routers, router)
	}
	return routers
}

func (srv *Server) Router(name string) *Router { return srv.routers[name] }

func (srv *Server) RemoveRouter(name string) {
	srv.group.Remove(name)
	delete(srv.routers, name)
}

func (router *Router) handle(path string, h HandlerFunc, filters []Filter, method ...string) {
	for i := len(filters) - 1; i >= 0; i-- {
		h = filters[i](h)
	}

	router.router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if ctx := router.srv.NewContext(w, r); ctx != nil { // NewContext 出错，则在 NewContext 中自行处理了输出内容。
			ctx.renderResponser(h(ctx))
			contextPool.Put(ctx)
		}
	}, method...)
}

func (router *Router) Handle(path string, h HandlerFunc, method ...string) *Router {
	router.handle(path, h, router.filters, method...)
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

// URL 构建完整的 URL
func (router *Router) URL(strict bool, pattern string, params map[string]string) (string, error) {
	return router.router.URL(strict, pattern, params)
}

// Prefix 返回特定前缀的路由设置对象
func (router *Router) Prefix(prefix string, filter ...Filter) *Prefix {
	filters := make([]Filter, 0, len(router.filters)+len(filter))
	filters = append(filters, router.filters...)
	filters = append(filters, filter...)
	return &Prefix{
		router:  router,
		prefix:  prefix,
		filters: filters,
	}
}

func (p *Prefix) Handle(path string, h HandlerFunc, method ...string) *Prefix {
	p.router.handle(p.prefix+path, h, p.filters, method...)
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
