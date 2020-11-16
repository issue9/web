// SPDX-License-Identifier: MIT

package context

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/issue9/mux/v3"
)

type (
	// HandlerFunc 路由项处理函数原型
	HandlerFunc func(*Context)

	// Prefix 带有统一前缀的路由管理
	Prefix interface {
		Handle(string, HandlerFunc, ...string) error
		Get(string, HandlerFunc) Prefix
		Put(string, HandlerFunc) Prefix
		Post(string, HandlerFunc) Prefix
		Patch(string, HandlerFunc) Prefix
		Delete(string, HandlerFunc) Prefix
		Options(string, string) Prefix
		Prefix(string, ...Filter) Prefix
		Resource(string, ...Filter) Resource
		Remove(string, ...string)
	}

	// Resource 同一资源的不同请求方法的管理
	Resource interface {
		Handle(HandlerFunc, ...string) error
		Get(HandlerFunc) Resource
		Put(HandlerFunc) Resource
		Post(HandlerFunc) Resource
		Patch(HandlerFunc) Resource
		Delete(HandlerFunc) Resource
		Options(string) Resource
		Remove(...string)
	}

	// Router 路由管理
	Router struct {
		srv *Server
		mux *mux.Mux

		root    string
		url     *url.URL
		filters []Filter
	}

	routerPrefix struct {
		srv *Server
		mux *mux.Mux

		prefix  string
		filters []Filter
	}

	routerResource struct {
		srv *Server
		mux *mux.Mux

		pattern string
		filters []Filter
	}

	modulePrefix struct {
		p       string
		m       *Module
		filters []Filter
	}

	moduleResource struct {
		m       *Module
		p       string
		filters []Filter
	}
)

func buildPrefix(srv *Server, mux *mux.Mux, prefix string, filter ...Filter) *routerPrefix {
	return &routerPrefix{
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
func (router *Router) Resource(pattern string, filter ...Filter) Resource {
	filters := make([]Filter, 0, len(router.filters)+len(filter))
	filters = append(filters, router.filters...)
	filters = append(filters, filter...)

	return &routerResource{
		srv: router.srv,
		mux: router.Mux(),

		pattern: router.url.Path + pattern,
		filters: filters,
	}
}

// Prefix 返回特定前缀的路由设置对象
func (router *Router) Prefix(prefix string, filter ...Filter) Prefix {
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
func (p *routerPrefix) Resource(pattern string, filter ...Filter) Resource {
	filters := make([]Filter, 0, len(p.filters)+len(filter))
	filters = append(filters, p.filters...)
	filters = append(filters, filter...)

	return &routerResource{
		srv: p.srv,
		mux: p.mux,

		pattern: p.prefix + pattern,
		filters: filters,
	}
}

// Prefix 返回特定前缀的路由设置对象
func (p *routerPrefix) Prefix(prefix string, filter ...Filter) Prefix {
	filters := make([]Filter, 0, len(p.filters)+len(filter))
	filters = append(filters, p.filters...)
	filters = append(filters, filter...)
	return buildPrefix(p.srv, p.mux, p.prefix+prefix, filters...)
}

// Handle 添加路由请求项
func (p *routerPrefix) Handle(path string, h HandlerFunc, method ...string) error {
	return p.srv.handle(p.mux, p.prefix+path, h, p.filters, method...)
}

// Handle 添加路由请求项
func (p *routerPrefix) handle(path string, h HandlerFunc, method ...string) Prefix {
	if err := p.Handle(path, h, method...); err != nil {
		panic(err)
	}
	return p
}

// Get 添加 GET 请求处理项
func (p *routerPrefix) Get(path string, h HandlerFunc) Prefix {
	return p.handle(path, h, http.MethodGet)
}

// Post 添加 POST 请求处理项
func (p *routerPrefix) Post(path string, h HandlerFunc) Prefix {
	return p.handle(path, h, http.MethodPost)
}

// Put 添加 PUT 请求处理项
func (p *routerPrefix) Put(path string, h HandlerFunc) Prefix {
	return p.handle(path, h, http.MethodPut)
}

// Delete 添加 DELETE 请求处理项
func (p *routerPrefix) Delete(path string, h HandlerFunc) Prefix {
	return p.handle(path, h, http.MethodDelete)
}

// Options 添加 OPTIONS 请求处理项
//
// 忽略 Filter 类型的是间件，如果有需要，可以采用 Handle 处理 Options 请求。
func (p *routerPrefix) Options(path, allow string) Prefix {
	p.mux.Options(p.prefix+path, allow)
	return p
}

// Patch 添加 PATCH 请求处理项
func (p *routerPrefix) Patch(path string, h HandlerFunc) Prefix {
	return p.handle(path, h, http.MethodPatch)
}

// Remove 删除指定的路由项
func (p *routerPrefix) Remove(path string, method ...string) {
	p.mux.Remove(p.prefix+path, method...)
}

// Handle 添加路由项
func (r *routerResource) Handle(h HandlerFunc, method ...string) error {
	return r.srv.handle(r.mux, r.pattern, h, r.filters, method...)
}

func (r *routerResource) handle(h HandlerFunc, method ...string) Resource {
	if err := r.Handle(h, method...); err != nil {
		panic(err)
	}
	return r
}

// Get 指定一个 GET 请求
func (r *routerResource) Get(h HandlerFunc) Resource {
	return r.handle(h, http.MethodGet)
}

// Post 指定个 POST 请求处理
func (r *routerResource) Post(h HandlerFunc) Resource {
	return r.handle(h, http.MethodPost)
}

// Delete 指定个 Delete 请求处理
func (r *routerResource) Delete(h HandlerFunc) Resource {
	return r.handle(h, http.MethodDelete)
}

// Put 指定个 Put 请求处理
func (r *routerResource) Put(h HandlerFunc) Resource {
	return r.handle(h, http.MethodPut)
}

// Patch 指定个 Patch 请求处理
func (r *routerResource) Patch(h HandlerFunc) Resource {
	return r.handle(h, http.MethodPatch)
}

// Remove 删除指定的路由项
func (r *routerResource) Remove(method ...string) {
	r.mux.Remove(r.pattern, method...)
}

///////////////////////// module

// AddFilters 添加过滤器
//
// 按给定参数的顺序反向依次调用。
func (m *Module) AddFilters(filter ...Filter) {
	m.filters = append(m.filters, filter...)
}

// Resource 生成资源项
func (m *Module) Resource(pattern string, filter ...Filter) Resource {
	return &moduleResource{
		m:       m,
		p:       pattern,
		filters: filter,
	}
}

// Prefix 声明一个 Prefix 实例
func (m *Module) Prefix(prefix string, filter ...Filter) Prefix {
	return &modulePrefix{
		m:       m,
		p:       prefix,
		filters: filter,
	}
}

// Handle 添加路由项
func (m *Module) Handle(path string, h HandlerFunc, method ...string) *Module {
	m.handle(path, h, nil, method...)
	return m
}

func (m *Module) handle(path string, h HandlerFunc, filter []Filter, method ...string) *Module {
	m.AddInit(func() error {
		filters := make([]Filter, len(m.filters)+len(filter))
		l := copy(filters, m.filters)
		copy(filters[l:], filter)

		h = FilterHandler(h, filters...)
		return m.srv.Router().Handle(path, h, method...)
	}, fmt.Sprintf("注册路由：[%s] %s", strings.Join(method, ","), path))
	return m
}

// Get 指定一个 GET 请求
func (m *Module) Get(path string, h HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodGet)
}

// Post 指定个 POST 请求处理
func (m *Module) Post(path string, h HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodPost)
}

// Delete 指定个 DELETE 请求处理
func (m *Module) Delete(path string, h HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodDelete)
}

// Put 指定个 PUT 请求处理
func (m *Module) Put(path string, h HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodPut)
}

// Patch 指定个 PATCH 请求处理
func (m *Module) Patch(path string, h HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodPatch)
}

// Options 指定个 OPTIONS 请求处理
func (m *Module) Options(path, allow string) *Module {
	m.AddInit(func() error {
		m.srv.Router().Options(path, allow)
		return nil
	}, fmt.Sprintf("注册路由：OPTIONS %s", path))
	return m
}

// Remove 删除指定的路由项
func (m *Module) Remove(path string, method ...string) {
	m.AddInit(func() error {
		m.srv.Router().Remove(path, method...)
		return nil
	}, fmt.Sprintf("删除路由项：%s", path))
}

// Options 添加 OPTIONS 请求处理项
//
// 忽略 Filter 类型的是间件，如果有需要，可以采用 Handle 处理 Options 请求。
func (r *routerResource) Options(allow string) Resource {
	r.mux.Options(r.pattern, allow)
	return r
}

// Resource 生成资源项
func (p *modulePrefix) Resource(pattern string, filter ...Filter) Resource {
	return p.m.Resource(p.p+pattern, filter...)
}

// Handle 添加路由项
func (r *moduleResource) Handle(h HandlerFunc, method ...string) error {
	r.m.handle(r.p, h, r.filters, method...)
	return nil
}

func (r *moduleResource) handle(h HandlerFunc, method ...string) Resource {
	r.m.handle(r.p, h, r.filters, method...)
	return r
}

// Get 指定一个 GET 请求
func (r *moduleResource) Get(h HandlerFunc) Resource {
	return r.handle(h, http.MethodGet)
}

// Post 指定个 POST 请求处理
func (r *moduleResource) Post(h HandlerFunc) Resource {
	return r.handle(h, http.MethodPost)
}

// Delete 指定个 DELETE 请求处理
func (r *moduleResource) Delete(h HandlerFunc) Resource {
	return r.handle(h, http.MethodDelete)
}

// Put 指定个 PUT 请求处理
func (r *moduleResource) Put(h HandlerFunc) Resource {
	return r.handle(h, http.MethodPut)
}

// Patch 指定个 PATCH 请求处理
func (r *moduleResource) Patch(h HandlerFunc) Resource {
	return r.handle(h, http.MethodPatch)
}

// Options 指定个 OPTIONS 请求处理
func (r *moduleResource) Options(allow string) Resource {
	r.m.Options(r.p, allow)
	return r
}

// Remove 删除指定的路由项
func (r *moduleResource) Remove(method ...string) {
	r.m.Remove(r.p, method...)
}

// Prefix 声明一个 Prefix 实例
func (p *modulePrefix) Prefix(prefix string, filter ...Filter) Prefix {
	return &modulePrefix{
		m:       p.m,
		p:       p.p + prefix,
		filters: filter,
	}
}

// Remove 删除指定的路由项
func (p *modulePrefix) Remove(path string, method ...string) {
	p.m.Remove(p.p+path, method...)
}

// Handle 添加路由项
func (p *modulePrefix) Handle(path string, h HandlerFunc, method ...string) error {
	p.m.handle(p.p+path, h, p.filters, method...)
	return nil
}

func (p *modulePrefix) handle(path string, h HandlerFunc, method ...string) Prefix {
	p.Handle(path, h, method...)
	return p
}

// Get 指定一个 GET 请求
func (p *modulePrefix) Get(path string, h HandlerFunc) Prefix {
	return p.handle(path, h, http.MethodGet)
}

// Post 指定个 POST 请求处理
func (p *modulePrefix) Post(path string, h HandlerFunc) Prefix {
	return p.handle(path, h, http.MethodPost)
}

// Delete 指定个 DELETE 请求处理
func (p *modulePrefix) Delete(path string, h HandlerFunc) Prefix {
	return p.handle(path, h, http.MethodDelete)
}

// Put 指定个 PUT 请求处理
func (p *modulePrefix) Put(path string, h HandlerFunc) Prefix {
	return p.handle(path, h, http.MethodPut)
}

// Patch 指定个 PATCH 请求处理
func (p *modulePrefix) Patch(path string, h HandlerFunc) Prefix {
	return p.handle(path, h, http.MethodPatch)
}

// Options 指定个 OPTIONS 请求处理
func (p *modulePrefix) Options(path, allow string) Prefix {
	p.m.Options(p.p+path, allow)
	return p
}
