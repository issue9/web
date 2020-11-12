// SPDX-License-Identifier: MIT

package context

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/issue9/mux/v3"
)

// HandlerFunc 路由项处理函数原型
type HandlerFunc func(*Context)

// Router 路由管理
type Router struct {
	srv *Server
	mux *mux.Mux

	root    string
	url     *url.URL
	filters []Filter
}

// Prefix 管理带有统一前缀的路由项
type Prefix struct {
	srv *Server
	mux *mux.Mux

	prefix  string
	filters []Filter
}

// Resource 以资源地址为对象的路由配置
type Resource struct {
	srv *Server
	mux *mux.Mux

	pattern string
	filters []Filter
}

func buildPrefix(srv *Server, mux *mux.Mux, prefix string, filter ...Filter) *Prefix {
	return &Prefix{
		srv: srv,
		mux: mux,

		prefix:  prefix,
		filters: filter,
	}
}

func buildRouter(srv *Server, mux *mux.Mux, u *url.URL, filter ...Filter) *Router {
	// 保证不以 / 结尾
	if len(u.Path) > 0 && u.Path[len(u.Path)-1] == '/' {
		u.Path = u.Path[:len(u.Path)-1]
	}

	return &Router{
		srv: srv,
		mux: mux,

		url:     u,
		root:    u.String(),
		filters: filter,
	}
}

// Router 返回操作路由项的实例
//
// 路由地址基于 root 的值，
// 且所有的路由都会自动应用通过 Server.AddFilters 和 Server.AddMiddlewares 添加的中间件。
func (srv *Server) Router() *Router {
	return srv.router
}

// handle
func (srv *Server) handle(mux *mux.Mux, path string, next HandlerFunc, filters []Filter, method ...string) error {
	return mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		fs := make([]Filter, 0, len(filters)+len(srv.filters))
		fs = append(fs, srv.filters...)
		fs = append(fs, filters...)
		FilterHandler(next, fs...)(srv.NewContext(w, r))
	}, method...)
}

// Mux 返回 mux.Mux 实例
func (router *Router) Mux() *mux.Mux {
	return router.mux
}

// Path 返回相对于域名的绝对路由地址
//
// 功能与 mux.URL 相似，但是加上了关联的域名地址的根路径。比如根地址是 https://example.com/blog
// pattern 为 /posts/{id}，则返回为 /blog/posts/1。
// 如果 params 为空的话，则会直接将 pattern 作为从 mux 转换之后的内容与 router.root 合并返回。
func (router *Router) Path(pattern string, params map[string]string) (string, error) {
	return router.buildURL(router.url.Path, pattern, params)
}

// URL 构建完整的 URL
//
// 功能与 mux.URL 相似，但是加上了关联的域名地址。比如根地址是 https://example.com/blog
// pattern 为 /posts/{id}，则返回为 https://example.com/blog/posts/1。
// 如果 params 为空的话，则会直接将 pattern 作为从 mux 转换之后的内容与 router.root 合并返回。
func (router *Router) URL(pattern string, params map[string]string) (string, error) {
	return router.buildURL(router.root, pattern, params)
}

func (router *Router) buildURL(prefix, pattern string, params map[string]string) (string, error) {
	if len(pattern) == 0 {
		return prefix, nil
	}

	if len(params) > 0 {
		p, err := router.Mux().URL(pattern, params)
		if err != nil {
			return "", err
		}
		pattern = p
	}

	switch {
	case pattern[0] == '/':
		// 由 buildRouter 保证 root 不能 / 结尾
		return prefix + pattern, nil
	default:
		return prefix + "/" + pattern, nil
	}
}

// NewRouter 构建基于 matcher 匹配的路由操作实例
//
// 不会应用通过 Server.AddFilters 添加的中间件，但是会应用 Server.AddMiddlewares 添加的中间件。
func (router *Router) NewRouter(name string, url *url.URL, matcher mux.Matcher, filter ...Filter) (*Router, bool) {
	m, ok := router.Mux().NewMux(name, matcher)
	if !ok {
		return nil, false
	}

	return buildRouter(router.srv, m, url, filter...), true
}

// Static 添加静态路由
//
// path 为路由地址，必须以命名参数结尾，比如 /assets/{path}，之后可以通过此值删除路由项；
// dir 为指向静态文件的路径；
//
// 如果要删除该静态路由，则可以将 path 传递给 Remove 进行删除。
//
// 比如在 Root 的值为 example.com/blog 时，
// 将参数指定为 /admin/{path} 和 ~/data/assets/admin
// 表示将 example.com/blog/admin/* 解析到 ~/data/assets/admin 目录之下。
func (router *Router) Static(p, dir string) error {
	p = path.Join(router.url.Path, p)
	if p != "" && p[0] != '/' {
		p = "/" + p
	}

	lastStart := strings.LastIndexByte(p, '{')
	if lastStart < 0 || len(p) == 0 || p[len(p)-1] != '}' || lastStart+2 == len(p) {
		return errors.New("path 必须是命名参数结尾：比如 /assets/{path}。")
	}

	h := http.StripPrefix(p[:lastStart], http.FileServer(http.Dir(dir)))
	return router.Mux().Handle(p, h, http.MethodGet)
}

// Resource 生成资源项
func (router *Router) Resource(pattern string, filter ...Filter) *Resource {
	filters := make([]Filter, 0, len(router.filters)+len(filter))
	filters = append(filters, router.filters...)
	filters = append(filters, filter...)

	return &Resource{
		srv: router.srv,
		mux: router.Mux(),

		pattern: router.url.Path + pattern,
		filters: filters,
	}
}

// Prefix 返回特定前缀的路由设置对象
func (router *Router) Prefix(prefix string, filter ...Filter) *Prefix {
	filters := make([]Filter, 0, len(router.filters)+len(filter))
	filters = append(filters, router.filters...)
	filters = append(filters, filter...)
	return buildPrefix(router.srv, router.mux, router.url.Path+prefix, filters...)
}

// Handle 添加路由请求项
func (router *Router) Handle(path string, h HandlerFunc, method ...string) error {
	return router.srv.handle(router.Mux(), router.url.Path+path, h, router.filters, method...)
}

// Handle 添加路由请求项
func (router *Router) handle(path string, h HandlerFunc, method ...string) *Router {
	if err := router.Handle(path, h, method...); err != nil {
		panic(err)
	}
	return router
}

// Get 添加 GET 请求处理项
func (router *Router) Get(path string, h HandlerFunc) *Router {
	return router.handle(path, h, http.MethodGet)
}

// Post 添加 POST 请求处理项
func (router *Router) Post(path string, h HandlerFunc) *Router {
	return router.handle(path, h, http.MethodPost)
}

// Put 添加 PUT 请求处理项
func (router *Router) Put(path string, h HandlerFunc) *Router {
	return router.handle(path, h, http.MethodPut)
}

// Delete 添加 DELETE 请求处理项
func (router *Router) Delete(path string, h HandlerFunc) *Router {
	return router.handle(path, h, http.MethodDelete)
}

// Options 添加 OPTIONS 请求处理项
//
// 忽略 Filter 类型的是间件，如果有需要，可以采用 Handle 处理 Options 请求。
func (router *Router) Options(path, allow string) *Router {
	router.Mux().Options(router.url.Path+path, allow)
	return router
}

// Patch 添加 PATCH 请求处理项
func (router *Router) Patch(path string, h HandlerFunc) *Router {
	return router.handle(path, h, http.MethodPatch)
}

// Remove 删除指定的路由项
func (router *Router) Remove(path string, method ...string) {
	router.Mux().Remove(router.url.Path+path, method...)
}

// Resource 生成资源项
func (p *Prefix) Resource(pattern string, filter ...Filter) *Resource {
	filters := make([]Filter, 0, len(p.filters)+len(filter))
	filters = append(filters, p.filters...)
	filters = append(filters, filter...)

	return &Resource{
		srv: p.srv,
		mux: p.mux,

		pattern: p.prefix + pattern,
		filters: filters,
	}
}

// Prefix 返回特定前缀的路由设置对象
func (p *Prefix) Prefix(prefix string, filter ...Filter) *Prefix {
	filters := make([]Filter, 0, len(p.filters)+len(filter))
	filters = append(filters, p.filters...)
	filters = append(filters, filter...)
	return buildPrefix(p.srv, p.mux, p.prefix+prefix, filters...)
}

// Handle 添加路由请求项
func (p *Prefix) Handle(path string, h HandlerFunc, method ...string) error {
	return p.srv.handle(p.mux, p.prefix+path, h, p.filters, method...)
}

// Handle 添加路由请求项
func (p *Prefix) handle(path string, h HandlerFunc, method ...string) *Prefix {
	if err := p.Handle(path, h, method...); err != nil {
		panic(err)
	}
	return p
}

// Get 添加 GET 请求处理项
func (p *Prefix) Get(path string, h HandlerFunc) *Prefix {
	return p.handle(path, h, http.MethodGet)
}

// Post 添加 POST 请求处理项
func (p *Prefix) Post(path string, h HandlerFunc) *Prefix {
	return p.handle(path, h, http.MethodPost)
}

// Put 添加 PUT 请求处理项
func (p *Prefix) Put(path string, h HandlerFunc) *Prefix {
	return p.handle(path, h, http.MethodPut)
}

// Delete 添加 DELETE 请求处理项
func (p *Prefix) Delete(path string, h HandlerFunc) *Prefix {
	return p.handle(path, h, http.MethodDelete)
}

// Options 添加 OPTIONS 请求处理项
//
// 忽略 Filter 类型的是间件，如果有需要，可以采用 Handle 处理 Options 请求。
func (p *Prefix) Options(path, allow string) *Prefix {
	p.mux.Options(p.prefix+path, allow)
	return p
}

// Patch 添加 PATCH 请求处理项
func (p *Prefix) Patch(path string, h HandlerFunc) *Prefix {
	return p.handle(path, h, http.MethodPatch)
}

// Remove 删除指定的路由项
func (p *Prefix) Remove(path string, method ...string) {
	p.mux.Remove(p.prefix+path, method...)
}

// Handle 添加路由项
func (r *Resource) Handle(h HandlerFunc, method ...string) error {
	return r.srv.handle(r.mux, r.pattern, h, r.filters, method...)
}

func (r *Resource) handle(h HandlerFunc, method ...string) *Resource {
	if err := r.Handle(h, method...); err != nil {
		panic(err)
	}
	return r
}

// Get 指定一个 GET 请求
func (r *Resource) Get(h HandlerFunc) *Resource {
	return r.handle(h, http.MethodGet)
}

// Post 指定个 POST 请求处理
func (r *Resource) Post(h HandlerFunc) *Resource {
	return r.handle(h, http.MethodPost)
}

// Delete 指定个 Delete 请求处理
func (r *Resource) Delete(h HandlerFunc) *Resource {
	return r.handle(h, http.MethodDelete)
}

// Put 指定个 Put 请求处理
func (r *Resource) Put(h HandlerFunc) *Resource {
	return r.handle(h, http.MethodPut)
}

// Patch 指定个 Patch 请求处理
func (r *Resource) Patch(h HandlerFunc) *Resource {
	return r.handle(h, http.MethodPatch)
}

// Remove 删除指定的路由项
func (r *Resource) Remove(method ...string) {
	r.mux.Remove(r.pattern, method...)
}

// Options 添加 OPTIONS 请求处理项
//
// 忽略 Filter 类型的是间件，如果有需要，可以采用 Handle 处理 Options 请求。
func (r *Resource) Options(allow string) *Resource {
	r.mux.Options(r.pattern, allow)
	return r
}
