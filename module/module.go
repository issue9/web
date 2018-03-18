// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package module 提供模块的的相关功能。
package module

import (
	"net/http"

	"github.com/issue9/web/core"
	"github.com/issue9/web/core/dependency"
)

// Module 表示模块信息
type Module struct {
	Name        string
	Deps        []string
	Description string
	Routes      []*Route

	inits      []dependency.InitFunc
	middleware core.Middleware
}

// Route 表示模块信息中的路由信息
type Route struct {
	Path    string
	Methods []string
	handler http.Handler
}

// Prefix 可以将具有统一前缀的路由项集中在一起操作。
//  p := srv.Prefix("/api")
//  p.Get("/users")  // 相当于 srv.Get("/api/users")
//  p.Get("/user/1") // 相当于 srv.Get("/api/user/1")
type Prefix struct {
	module *Module
	prefix string
}

// New 声明一个新的模块
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func (ms *Modules) New(name, desc string, deps ...string) *Module {
	m := &Module{
		Name:        name,
		Deps:        deps,
		Description: desc,
		Routes:      make([]*Route, 0, 10),
		inits:       make([]dependency.InitFunc, 0, 5),
	}

	ms.modules = append(ms.modules, m)
	return m
}

// Prefix 声明一个 Prefix 实例。
func (m *Module) Prefix(prefix string) *Prefix {
	return &Prefix{
		module: m,
		prefix: prefix,
	}
}

// Module 返回关联的 Module 实全
func (p *Prefix) Module() *Module {
	return p.module
}

// SetMiddleware 为当前模块的所有路由添加的中间件。
//
// 多次调用，最后一次为准。
func (m *Module) SetMiddleware(f core.Middleware) {
	m.middleware = f
}

// AddInit 添加一个初始化函数
func (m *Module) AddInit(f dependency.InitFunc) *Module {
	m.inits = append(m.inits, f)
	return m
}

// Handle 添加一个路由项
func (m *Module) Handle(path string, h http.Handler, methods ...string) *Module {
	m.Routes = append(m.Routes, &Route{
		Methods: methods,
		Path:    path,
		handler: h,
	})

	return m
}

// Get 指定一个 GET 请求
func (m *Module) Get(path string, h http.Handler) *Module {
	return m.Handle(path, h, http.MethodGet)
}

// Post 指定个 POST 请求处理
func (m *Module) Post(path string, h http.Handler) *Module {
	return m.Handle(path, h, http.MethodPost)
}

// Delete 指定个 Delete 请求处理
func (m *Module) Delete(path string, h http.Handler) *Module {
	return m.Handle(path, h, http.MethodDelete)
}

// Put 指定个 Put 请求处理
func (m *Module) Put(path string, h http.Handler) *Module {
	return m.Handle(path, h, http.MethodPut)
}

// Patch 指定个 Patch 请求处理
func (m *Module) Patch(path string, h http.Handler) *Module {
	return m.Handle(path, h, http.MethodPatch)
}

// HandleFunc 指定一个请求
func (m *Module) HandleFunc(path string, h func(w http.ResponseWriter, r *http.Request), methods ...string) *Module {
	return m.Handle(path, http.HandlerFunc(h), methods...)
}

// GetFunc 指定一个 GET 请求
func (m *Module) GetFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.HandleFunc(path, h, http.MethodGet)
}

// PostFunc 指定一个 Post 请求
func (m *Module) PostFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.HandleFunc(path, h, http.MethodPost)
}

// DeleteFunc 指定一个 Delete 请求
func (m *Module) DeleteFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.HandleFunc(path, h, http.MethodDelete)
}

// PutFunc 指定一个 Put 请求
func (m *Module) PutFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.HandleFunc(path, h, http.MethodPut)
}

// PatchFunc 指定一个 Patch 请求
func (m *Module) PatchFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.HandleFunc(path, h, http.MethodPatch)
}

// Handle 添加路由项
func (p *Prefix) Handle(path string, h http.Handler, methods ...string) *Prefix {
	p.module.Handle(p.prefix+path, h, methods...)
	return p
}

// Get 指定一个 GET 请求
func (p *Prefix) Get(path string, h http.Handler) *Prefix {
	return p.Handle(path, h, http.MethodGet)
}

// Post 指定一个 Post 请求
func (p *Prefix) Post(path string, h http.Handler) *Prefix {
	return p.Handle(path, h, http.MethodPost)
}

// Delete 指定一个 Delete 请求
func (p *Prefix) Delete(path string, h http.Handler) *Prefix {
	return p.Handle(path, h, http.MethodDelete)
}

// Put 指定一个 Put 请求
func (p *Prefix) Put(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodPut)
}

// Patch 指定一个 Patch 请求
func (p *Prefix) Patch(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodPatch)
}

// HandleFunc 指定一个请求
func (p *Prefix) HandleFunc(path string, f http.HandlerFunc, methods ...string) *Prefix {
	return p.Handle(path, http.HandlerFunc(f), methods...)
}

// GetFunc 指定一个 Get 请求
func (p *Prefix) GetFunc(path string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(path, f, http.MethodGet)
}

// PutFunc 指定一个 Put 请求
func (p *Prefix) PutFunc(path string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(path, f, http.MethodPut)
}

// PostFunc 指定一个 Post 请求
func (p *Prefix) PostFunc(path string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(path, f, http.MethodPost)
}

// DeleteFunc 指定一个 Delete 请求
func (p *Prefix) DeleteFunc(path string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(path, f, http.MethodDelete)
}

// PatchFunc 指定一个 Patch 请求
func (p *Prefix) PatchFunc(path string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(path, f, http.MethodPatch)
}
