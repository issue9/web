// SPDX-License-Identifier: MIT

package module

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/issue9/web/context"
)

// Prefix 管理带有统一前缀的路由项
type Prefix struct {
	p string
	m *Module
}

// Resource 以资源地址为对象的路由配置
type Resource struct {
	m *Module
	p string
}

// Resource 生成资源项
func (m *Module) Resource(pattern string) *Resource {
	return &Resource{m: m, p: pattern}
}

// Resource 生成资源项
func (p *Prefix) Resource(pattern string) *Resource {
	return p.m.Resource(p.p + pattern)
}

// Handle 添加路由项
func (r *Resource) Handle(h context.HandlerFunc, method ...string) *Resource {
	r.m.Handle(r.p, h, method...)
	return r
}

// Get 指定一个 GET 请求
func (r *Resource) Get(h context.HandlerFunc) *Resource {
	return r.Handle(h, http.MethodGet)
}

// Post 指定个 POST 请求处理
func (r *Resource) Post(h context.HandlerFunc) *Resource {
	return r.Handle(h, http.MethodPost)
}

// Delete 指定个 DELETE 请求处理
func (r *Resource) Delete(h context.HandlerFunc) *Resource {
	return r.Handle(h, http.MethodDelete)
}

// Put 指定个 PUT 请求处理
func (r *Resource) Put(h context.HandlerFunc) *Resource {
	return r.Handle(h, http.MethodPut)
}

// Patch 指定个 PATCH 请求处理
func (r *Resource) Patch(h context.HandlerFunc) *Resource {
	return r.Handle(h, http.MethodPatch)
}

// Options 指定个 OPTIONS 请求处理
func (r *Resource) Options(allow string) *Resource {
	r.m.Options(r.p, allow)
	return r
}

// Module 返回关联的模块实例
func (p *Prefix) Module() *Module {
	return p.m
}

// Handle 添加路由项
func (p *Prefix) Handle(path string, h context.HandlerFunc, method ...string) *Prefix {
	p.Module().Handle(p.p+path, h, method...)
	return p
}

// Get 指定一个 GET 请求
func (p *Prefix) Get(path string, h context.HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodGet)
}

// Post 指定个 POST 请求处理
func (p *Prefix) Post(path string, h context.HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodPost)
}

// Delete 指定个 DELETE 请求处理
func (p *Prefix) Delete(path string, h context.HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodDelete)
}

// Put 指定个 PUT 请求处理
func (p *Prefix) Put(path string, h context.HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodPut)
}

// Patch 指定个 PATCH 请求处理
func (p *Prefix) Patch(path string, h context.HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodPatch)
}

// Options 指定个 OPTIONS 请求处理
func (p *Prefix) Options(path, allow string) *Prefix {
	p.m.Options(p.p+path, allow)
	return p
}

// Prefix 声明一个 Prefix 实例
func (m *Module) Prefix(prefix string) *Prefix {
	return &Prefix{m: m, p: prefix}
}

// Handle 添加路由项
func (m *Module) Handle(path string, h context.HandlerFunc, method ...string) *Module {
	m.AddInit(func() error {
		return m.srv.ctxServer.Handle(path, h, method...)
	}, fmt.Sprintf("注册路由：[%s] %s", strings.Join(method, ","), path))
	return m
}

// Get 指定一个 GET 请求
func (m *Module) Get(path string, h context.HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodGet)
}

// Post 指定个 POST 请求处理
func (m *Module) Post(path string, h context.HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodPost)
}

// Delete 指定个 DELETE 请求处理
func (m *Module) Delete(path string, h context.HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodDelete)
}

// Put 指定个 PUT 请求处理
func (m *Module) Put(path string, h context.HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodPut)
}

// Patch 指定个 PATCH 请求处理
func (m *Module) Patch(path string, h context.HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodPatch)
}

// Options 指定个 OPTIONS 请求处理
func (m *Module) Options(path, allow string) *Module {
	m.AddInit(func() error {
		m.srv.ctxServer.Options(path, allow)
		return nil
	}, fmt.Sprintf("注册路由：OPTIONS %s", path))
	return m
}
