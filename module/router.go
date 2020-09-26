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

// Delete 指定个 Delete 请求处理
func (p *Prefix) Delete(path string, h context.HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodDelete)
}

// Put 指定个 Put 请求处理
func (p *Prefix) Put(path string, h context.HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodPut)
}

// Patch 指定个 Patch 请求处理
func (p *Prefix) Patch(path string, h context.HandlerFunc) *Prefix {
	return p.Handle(path, h, http.MethodPatch)
}

// Prefix 声明一个 Prefix 实例
func (m *Module) Prefix(prefix string) *Prefix {
	return &Prefix{m: m, p: prefix}
}

// Handle 添加路由项
func (m *Module) Handle(path string, h context.HandlerFunc, methods ...string) *Module {
	m.AddInit(func() error {
		return m.srv.ctxServer.Handle(path, h, methods...)
	}, fmt.Sprintf("注册路由：[%s] %s", strings.Join(methods, ","), path))
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

// Delete 指定个 Delete 请求处理
func (m *Module) Delete(path string, h context.HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodDelete)
}

// Put 指定个 Put 请求处理
func (m *Module) Put(path string, h context.HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodPut)
}

// Patch 指定个 Patch 请求处理
func (m *Module) Patch(path string, h context.HandlerFunc) *Module {
	return m.Handle(path, h, http.MethodPatch)
}
