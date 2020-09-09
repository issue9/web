// SPDX-License-Identifier: MIT

package server

import "github.com/issue9/web/context"

// Prefix 声明一个 Prefix 实例。
func (m *Module) Prefix(prefix string) *context.Prefix {
	return m.srv.builder.Prefix(prefix)
}

// Handle 添加一个路由项
func (m *Module) Handle(path string, h context.HandlerFunc, methods ...string) error {
	return m.srv.builder.Handle(path, h, methods...)
}

// Get 指定一个 GET 请求
func (m *Module) Get(path string, h context.HandlerFunc) *Module {
	m.srv.builder.Get(path, h)
	return m
}

// Post 指定个 POST 请求处理
func (m *Module) Post(path string, h context.HandlerFunc) *Module {
	m.srv.builder.Post(path, h)
	return m
}

// Delete 指定个 Delete 请求处理
func (m *Module) Delete(path string, h context.HandlerFunc) *Module {
	m.srv.builder.Delete(path, h)
	return m
}

// Put 指定个 Put 请求处理
func (m *Module) Put(path string, h context.HandlerFunc) *Module {
	m.srv.builder.Put(path, h)
	return m
}

// Patch 指定个 Patch 请求处理
func (m *Module) Patch(path string, h context.HandlerFunc) *Module {
	m.srv.builder.Patch(path, h)
	return m
}
