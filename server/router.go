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

	"github.com/issue9/middleware/v5/debugger"
	"github.com/issue9/mux/v5"
	"github.com/issue9/mux/v5/group"
)

type (
	// Router 路由管理
	Router struct {
		srv      *Server
		router   *mux.Router
		filters  []Filter
		root     string
		path     string
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
func (srv *Server) NewRouter(name string, root string, matcher group.Matcher, filter ...Filter) (*Router, error) {
	u, err := url.Parse(root)
	if err != nil {
		return nil, err
	}

	// 保证不以 / 结尾
	if len(u.Path) > 0 && u.Path[len(u.Path)-1] == '/' {
		u.Path = u.Path[:len(u.Path)-1]
		root = u.String()
	}

	r := srv.MuxGroups().NewRouter(name, matcher)
	dbg := &debugger.Debugger{}
	r.Middlewares().Append(dbg.Middleware)
	rr := &Router{
		srv:      srv,
		router:   r,
		filters:  filter,
		root:     root,
		path:     u.Path,
		debugger: dbg,
	}
	srv.routers[name] = rr

	return rr, nil
}

// Router 返回由 Server.NewRouter 声明的路由
func (srv *Server) Router(name string) *Router { return srv.routers[name] }

func (srv *Server) RemoveRouter(name string) {
	srv.MuxGroups().RemoveRouter(name)
	delete(srv.routers, name)
}

// MuxGroups 返回 group.Groups 实例
func (srv *Server) MuxGroups() *group.Groups { return srv.groups }

// SetDebugger 设置调试地址
func (router *Router) SetDebugger(pprof, vars string) (err error) {
	if pprof != "" {
		if pprof, err = router.Path(pprof, nil); err != nil {
			return err
		}
	}

	if vars != "" {
		if vars, err = router.Path(vars, nil); err != nil {
			return err
		}
	}

	router.debugger.Pprof = pprof
	router.debugger.Vars = vars

	return nil
}

func (router *Router) handleWithFilters(path string, h HandlerFunc, filters []Filter, method ...string) {
	h = ApplyFilters(h, filters...)
	router.router.HandleFunc(router.path+path, func(w http.ResponseWriter, r *http.Request) {
		ctx := router.srv.NewContext(w, r)
		ctx.renderResponser(h(ctx))
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

// Post 添加 POST 请求处理项
func (router *Router) Post(path string, h HandlerFunc) *Router {
	return router.Handle(path, h, http.MethodPost)
}

// Put 添加 PUT 请求处理项
func (router *Router) Put(path string, h HandlerFunc) *Router {
	return router.Handle(path, h, http.MethodPut)
}

// Delete 添加 DELETE 请求处理项
func (router *Router) Delete(path string, h HandlerFunc) *Router {
	return router.Handle(path, h, http.MethodDelete)
}

// Patch 添加 PATCH 请求处理项
func (router *Router) Patch(path string, h HandlerFunc) *Router {
	return router.Handle(path, h, http.MethodPatch)
}

// Remove 删除指定的路由项
func (router *Router) Remove(path string, method ...string) {
	router.router.Remove(router.path+path, method...)
}

func (router *Router) clone() *Router {
	filters := make([]Filter, len(router.filters))
	copy(filters, router.filters)
	return &Router{
		srv:     router.srv,
		router:  router.router,
		filters: filters,
		root:    router.root,
		path:    router.path,
	}
}

// Path 返回相对于域名的绝对路由地址
//
// 功能与 mux.Router.URL 相似，但是加上了关联的域名地址的根路径。比如根地址是 https://example.com/blog
// pattern 为 /posts/{id}，则返回为 /blog/posts/1。
// 如果 params 为空的话，则会直接将 pattern 作为从 mux 转换之后的内容与 router.root 合并返回。
func (router *Router) Path(pattern string, params map[string]string) (string, error) {
	return router.buildURL(router.path, pattern, params)
}

// URL 构建完整的 URL
//
// 功能与 mux.Router.URL 相似，但是加上了关联的域名地址。比如根地址是 https://example.com/blog
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
		p, err := router.router.URL(pattern, params)
		if err != nil {
			return "", err
		}
		pattern = p
	}

	switch {
	case pattern[0] == '/':
		// 由 NewRouter 保证 root 不能 / 结尾
		return prefix + pattern, nil
	default:
		return prefix + "/" + pattern, nil
	}
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
func (router *Router) Static(p, dir, index string) error {
	return router.StaticFS(p, os.DirFS(dir), index)
}

func (router *Router) StaticFS(p string, f fs.FS, index string) error {
	lastStart := strings.LastIndexByte(p, '{')
	if lastStart < 0 || len(p) == 0 || p[len(p)-1] != '}' || lastStart+2 == len(p) {
		return errors.New("path 必须是命名参数结尾：比如 /assets/{path}。")
	}
	prefix := path.Join(router.path, p[:lastStart])

	router.Handle(p, func(ctx *Context) Responser {
		pp := ctx.Request.URL.Path
		pp = strings.TrimPrefix(pp, prefix)
		if pp != "" && pp[0] == '/' {
			pp = pp[1:]
		}
		return ctx.ServeFileFS(f, pp, index, nil)
	}, http.MethodGet)
	return nil
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

// Get 添加 GET 请求处理项
func (p *Prefix) Get(path string, h HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodGet)
}

// Post 添加 POST 请求处理项
func (p *Prefix) Post(path string, h HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodPost)
}

// Put 添加 PUT 请求处理项
func (p *Prefix) Put(path string, h HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodPut)
}

// Delete 添加 DELETE 请求处理项
func (p *Prefix) Delete(path string, h HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodDelete)
}

// Patch 添加 Patch 请求处理
func (p *Prefix) Patch(path string, h HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodPatch)
}

// Remove 删除路由项
func (p *Prefix) Remove(path string, method ...string) {
	p.router.Remove(p.prefix+path, method...)
}

// AddRoutes 注册路由项
//
// f 实际执行注册路由的函数；
// routerName 路由名称，由 Server.NewRouter 中创建，若为空值，则采用 Action.Name 作为默认值；
func (t *Action) AddRoutes(f func(r *Router), routerName string) *Action {
	if routerName == "" {
		routerName = t.Name()
	}

	msg := t.Server().LocalePrinter().Sprintf("register router %s", routerName)
	return t.AddInit(msg, func() error {
		r := t.Server().Router(routerName)
		if r == nil {
			return fmt.Errorf("路由名 %s 不存在", routerName)
		}

		f(r)
		return nil
	})
}
