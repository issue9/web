// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/issue9/mux/v3"
)

type (
	// HandlerFunc 路由项处理函数原型
	HandlerFunc func(*Context)

	// Router 路由管理
	Router struct {
		srv *Server
		mux *mux.Mux

		root    string
		url     *url.URL
		filters []Filter
	}

	// Prefix 带有统一前缀的路由管理
	Prefix struct {
		handler       func(path string, h HandlerFunc, filters []Filter, method ...string) error
		optionHandler func(path string, allow string)

		srv     *Server
		mux     *mux.Mux
		prefix  string
		filters []Filter
	}

	// Resource 同一资源的不同请求方法的管理
	Resource struct {
		handler       func(h HandlerFunc, filters []Filter, method ...string) error
		optionHandler func(allow string)

		srv     *Server
		mux     *mux.Mux
		pattern string
		filters []Filter
	}
)

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
// p 为路由地址，必须以命名参数结尾，比如 /assets/{path}，之后可以通过此值删除路由项；
// index 可以在访问一个目录时指定默认访问的页面。
// dir 为指向静态文件的路径；
//
// 如果要删除该静态路由，则可以将 path 传递给 Remove 进行删除。
//
// 比如在 Root 的值为 example.com/blog 时，
// 将参数指定为 /admin/{path} 和 ~/data/assets/admin
// 表示将 example.com/blog/admin/* 解析到 ~/data/assets/admin 目录之下。
func (router *Router) Static(p, dir, index string) error {
	return router.StaticFS(p, os.DirFS(dir), index)
}

func (router *Router) StaticFS(p string, f fs.FS, index string) error {
	lastStart := strings.LastIndexByte(p, '{')
	if lastStart < 0 || len(p) == 0 || p[len(p)-1] != '}' || lastStart+2 == len(p) {
		return errors.New("path 必须是命名参数结尾：比如 /assets/{path}。")
	}
	prefix := path.Join(router.url.Path, p[:lastStart])

	return router.Handle(p, func(ctx *Context) {
		pp := ctx.Request.URL.Path
		pp = strings.TrimPrefix(pp, prefix)
		if pp != "" && pp[0] == '/' {
			pp = pp[1:]
		}
		ctx.ServeFileFS(f, pp, index, nil)
	}, http.MethodGet)
}

// Resource 生成资源项
func (router *Router) Resource(pattern string, filter ...Filter) *Resource {
	filters := make([]Filter, 0, len(router.filters)+len(filter))
	filters = append(filters, router.filters...)
	filters = append(filters, filter...)

	pattern = router.url.Path + pattern
	return &Resource{
		handler: func(h HandlerFunc, filters []Filter, method ...string) error {
			return router.srv.handle(router.mux, pattern, h, filters, method...)
		},
		optionHandler: func(allow string) {
			router.mux.Options(pattern, allow)
		},

		srv:     router.srv,
		mux:     router.Mux(),
		pattern: pattern,
		filters: filters,
	}
}

// Prefix 返回特定前缀的路由设置对象
func (router *Router) Prefix(prefix string, filter ...Filter) *Prefix {
	filters := make([]Filter, 0, len(router.filters)+len(filter))
	filters = append(filters, router.filters...)
	filters = append(filters, filter...)

	return &Prefix{
		handler: func(path string, h HandlerFunc, filters []Filter, method ...string) error {
			return router.srv.handle(router.mux, path, h, filters, method...)
		},
		optionHandler: func(path, allow string) {
			router.mux.Options(path, allow)
		},

		srv:     router.srv,
		mux:     router.mux,
		prefix:  router.url.Path + prefix,
		filters: filters,
	}
}

// Handle 添加路由请求项
func (router *Router) Handle(path string, h HandlerFunc, method ...string) error {
	return router.srv.handle(router.Mux(), router.url.Path+path, h, router.filters, method...)
}

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

// Resource 生成 Resource 对象
func (p *Prefix) Resource(pattern string, filter ...Filter) *Resource {
	filters := make([]Filter, 0, len(p.filters)+len(filter))
	filters = append(filters, p.filters...)
	filters = append(filters, filter...)

	pattern = p.prefix + pattern
	return &Resource{
		handler: func(h HandlerFunc, filters []Filter, method ...string) error {
			return p.handler(pattern, h, filters, method...)
		},
		optionHandler: func(allow string) {
			p.optionHandler(pattern, allow)
		},

		srv:     p.srv,
		mux:     p.mux,
		pattern: pattern,
		filters: filters,
	}
}

// Handle 添加路由项
func (p *Prefix) Handle(path string, h HandlerFunc, method ...string) error {
	return p.handler(p.prefix+path, h, p.filters, method...)
}

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

// Options 指定 OPTIONS 请求的返回内容
func (p *Prefix) Options(path, allow string) *Prefix {
	p.optionHandler(p.prefix+path, allow)
	return p
}

// Patch 添加 Patch 请求处理
func (p *Prefix) Patch(path string, h HandlerFunc) *Prefix {
	return p.handle(path, h, http.MethodPatch)
}

// Remove 删除路由项
func (p *Prefix) Remove(path string, method ...string) {
	p.mux.Remove(p.prefix+path, method...)
}

// Handle 添加路由项
func (r *Resource) Handle(h HandlerFunc, method ...string) error {
	return r.handler(h, r.filters, method...)
}

func (r *Resource) handle(h HandlerFunc, method ...string) *Resource {
	if err := r.Handle(h, method...); err != nil {
		panic(err)
	}
	return r
}

// Get 添加 GET 请求处理项
func (r *Resource) Get(h HandlerFunc) *Resource {
	return r.handle(h, http.MethodGet)
}

// Post 添加 POST 请求处理项
func (r *Resource) Post(h HandlerFunc) *Resource {
	return r.handle(h, http.MethodPost)
}

// Delete 添加 DELETE 请求处理项
func (r *Resource) Delete(h HandlerFunc) *Resource {
	return r.handle(h, http.MethodDelete)
}

// Put 添加 PUT 请求处理项
func (r *Resource) Put(h HandlerFunc) *Resource {
	return r.handle(h, http.MethodPut)
}

// Patch 添加 PATCH 请求处理项
func (r *Resource) Patch(h HandlerFunc) *Resource {
	return r.handle(h, http.MethodPatch)
}

// Remove 删除路由项
func (r *Resource) Remove(method ...string) {
	r.mux.Remove(r.pattern, method...)
}

// Options 指定 OPTIONS 请求的返回内容
func (r *Resource) Options(allow string) *Resource {
	r.optionHandler(allow)
	return r
}

// Handle 添加路由项
func (m *Module) Handle(path string, h HandlerFunc, method ...string) *Module {
	m.handle(path, h, nil, method...)
	return m
}

func (m *Module) handle(path string, h HandlerFunc, filter []Filter, method ...string) *Module {
	m.AddInit(fmt.Sprintf("注册路由：[%s] %s", strings.Join(method, ","), path), func() error {
		filters := make([]Filter, len(m.filters)+len(filter))
		l := copy(filters, m.filters)
		copy(filters[l:], filter)

		h = FilterHandler(h, filters...)
		return m.srv.Router().Handle(path, h, method...)
	})
	return m
}

// Get 添加 GET 请求处理项
func (m *Module) Get(path string, h HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodGet)
}

// Post 添加 POST 请求处理项
func (m *Module) Post(path string, h HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodPost)
}

// Delete 添加 DELETE 请求处理项
func (m *Module) Delete(path string, h HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodDelete)
}

// Put 添加 PUT 请求处理项
func (m *Module) Put(path string, h HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodPut)
}

// Patch 添加 Patch 请求处理
func (m *Module) Patch(path string, h HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodPatch)
}

// Options 指定 OPTIONS 请求的返回内容
func (m *Module) Options(path, allow string) *Module {
	m.AddInit(fmt.Sprintf("注册路由：OPTIONS %s", path), func() error {
		m.srv.Router().Options(path, allow)
		return nil
	})
	return m
}

// Remove 删除路由项
func (m *Module) Remove(path string, method ...string) *Module {
	m.AddInit(fmt.Sprintf("删除路由项：%s", path), func() error {
		m.srv.Router().Remove(path, method...)
		return nil
	})
	return m
}

// Prefix 返回特定前缀的路由设置对象
func (m *Module) Prefix(prefix string, filter ...Filter) *Prefix {
	return &Prefix{
		handler: func(path string, h HandlerFunc, filters []Filter, method ...string) error {
			m.handle(path, h, filters, method...)
			return nil
		},
		optionHandler: func(path, allow string) {
			m.Options(path, allow)
		},

		prefix:  prefix,
		filters: filter,
	}
}

// Resource 生成 Resource 对象
func (m *Module) Resource(pattern string, filter ...Filter) *Resource {
	return &Resource{
		handler: func(h HandlerFunc, filters []Filter, method ...string) error {
			m.handle(pattern, h, filters, method...)
			return nil
		},
		optionHandler: func(allow string) {
			m.Options(pattern, allow)
		},

		srv:     m.srv,
		mux:     m.srv.router.mux,
		pattern: pattern,
		filters: filter,
	}
}
