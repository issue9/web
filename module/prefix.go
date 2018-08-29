// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import "net/http"

// Prefix 可以将具有统一前缀的路由项集中在一起操作。
//  p := srv.Prefix("/api")
//  p.Get("/users")  // 相当于 srv.Get("/api/users")
//  p.Get("/user/1") // 相当于 srv.Get("/api/user/1")
type Prefix struct {
	module *Module
	prefix string
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
