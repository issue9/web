// SPDX-License-Identifier: MIT

package context

import (
	"net/http"

	"github.com/issue9/mux/v2"
)

// HandlerFunc 路由项处理函数原型
type HandlerFunc func(*Context)

// Prefix 管理带有统一前缀的路由项
type Prefix struct {
	prefix string
	srv    *Server
}

// Resource 以资源地址为对象的路由配置
type Resource struct {
	srv     *Server
	pattern string
}

// Resource 生成资源项
func (srv *Server) Resource(pattern string) *Resource {
	return &Resource{srv: srv, pattern: pattern}
}

// Resource 生成资源项
func (p *Prefix) Resource(pattern string) *Resource {
	return &Resource{srv: p.srv, pattern: p.prefix + pattern}
}

// Handle 添加路由项
func (r *Resource) Handle(h HandlerFunc, method ...string) *Resource {
	r.srv.Handle(r.pattern, h, method...)
	return r
}

// Get 指定一个 GET 请求
func (r *Resource) Get(h HandlerFunc) *Resource {
	return r.Handle(h, http.MethodGet)
}

// Post 指定个 POST 请求处理
func (r *Resource) Post(h HandlerFunc) *Resource {
	return r.Handle(h, http.MethodPost)
}

// Delete 指定个 Delete 请求处理
func (r *Resource) Delete(h HandlerFunc) *Resource {
	return r.Handle(h, http.MethodDelete)
}

// Put 指定个 Put 请求处理
func (r *Resource) Put(h HandlerFunc) *Resource {
	return r.Handle(h, http.MethodPut)
}

// Patch 指定个 Patch 请求处理
func (r *Resource) Patch(h HandlerFunc) *Resource {
	return r.Handle(h, http.MethodPatch)
}

// Remove 删除指定的路由项
func (r *Resource) Remove(method ...string) {
	r.srv.Remove(r.pattern, method...)
}

// Options 添加 OPTIONS 请求处理项
func (r *Resource) Options(allow string) *Resource {
	r.srv.Options(r.pattern, allow)
	return r
}

// Router 返回操作路由项的实例
//
// Router 可以处理兼容标准库的 net/http.Handler。
func (srv *Server) Router() *mux.Prefix {
	return srv.router
}

// Handle 添加路由请求项
func (srv *Server) Handle(path string, h HandlerFunc, method ...string) error {
	return srv.router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		h(srv.newContext(w, r))
	}, method...)
}

// Remove 删除指定的路由项
func (srv *Server) Remove(path string, method ...string) {
	srv.router.Remove(path, method...)
}

// Handle 添加路由请求项
func (srv *Server) handle(path string, h HandlerFunc, method ...string) *Server {
	if err := srv.Handle(path, h, method...); err != nil { // 路由项语法错误，直接 panic
		panic(err)
	}
	return srv
}

// Get 添加 GET 请求处理项
func (srv *Server) Get(path string, h HandlerFunc) *Server {
	return srv.handle(path, h, http.MethodGet)
}

// Post 添加 POST 请求处理项
func (srv *Server) Post(path string, h HandlerFunc) *Server {
	return srv.handle(path, h, http.MethodPost)
}

// Put 添加 PUT 请求处理项
func (srv *Server) Put(path string, h HandlerFunc) *Server {
	return srv.handle(path, h, http.MethodPut)
}

// Delete 添加 DELETE 请求处理项
func (srv *Server) Delete(path string, h HandlerFunc) *Server {
	return srv.handle(path, h, http.MethodDelete)
}

// Options 添加 OPTIONS 请求处理项
func (srv *Server) Options(path, allow string) *Server {
	srv.router.Options(path, allow)
	return srv
}

// Patch 添加 PATCH 请求处理项
func (srv *Server) Patch(path string, h HandlerFunc) *Server {
	return srv.handle(path, h, http.MethodPatch)
}

// Prefix 返回特定前缀的路由设置对象
func (srv *Server) Prefix(prefix string) *Prefix {
	return &Prefix{
		prefix: prefix,
		srv:    srv,
	}
}

// Handle 添加路由请求项
func (p *Prefix) Handle(path string, h HandlerFunc, method ...string) error {
	return p.srv.Handle(p.prefix+path, h, method...)
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
func (p *Prefix) Options(path, allow string) *Prefix {
	p.srv.Options(p.prefix+path, allow)
	return p
}

// Patch 添加 PATCH 请求处理项
func (p *Prefix) Patch(path string, h HandlerFunc) *Prefix {
	return p.handle(path, h, http.MethodPatch)
}

// Remove 删除指定的路由项
func (p *Prefix) Remove(path string, method ...string) {
	p.srv.Remove(p.prefix+path, method...)
}
