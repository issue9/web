// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"

	"github.com/issue9/mux/v2"
)

// Prefix 声明一个 Prefix 实例。
func (m *Module) Prefix(prefix string) *mux.Prefix {
	return m.app.router.Prefix(prefix)
}

// Handle 添加一个路由项
func (m *Module) Handle(path string, h http.Handler, methods ...string) error {
	return m.app.router.Handle(path, h, methods...)
}

// Get 指定一个 GET 请求
func (m *Module) Get(path string, h http.Handler) *mux.Prefix {
	return m.app.router.Get(path, h)
}

// Post 指定个 POST 请求处理
func (m *Module) Post(path string, h http.Handler) *mux.Prefix {
	return m.app.router.Post(path, h)
}

// Delete 指定个 Delete 请求处理
func (m *Module) Delete(path string, h http.Handler) *mux.Prefix {
	return m.app.router.Delete(path, h)
}

// Put 指定个 Put 请求处理
func (m *Module) Put(path string, h http.Handler) *mux.Prefix {
	return m.app.router.Put(path, h)
}

// Patch 指定个 Patch 请求处理
func (m *Module) Patch(path string, h http.Handler) *mux.Prefix {
	return m.app.router.Patch(path, h)
}

// HandleFunc 指定一个请求
func (m *Module) HandleFunc(path string, h func(w http.ResponseWriter, r *http.Request), methods ...string) error {
	return m.app.router.HandleFunc(path, h, methods...)
}

// GetFunc 指定一个 GET 请求
func (m *Module) GetFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return m.app.router.GetFunc(path, h)
}

// PostFunc 指定一个 Post 请求
func (m *Module) PostFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return m.app.router.PostFunc(path, h)
}

// DeleteFunc 指定一个 Delete 请求
func (m *Module) DeleteFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return m.app.router.DeleteFunc(path, h)
}

// PutFunc 指定一个 Put 请求
func (m *Module) PutFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return m.app.router.PutFunc(path, h)
}

// PatchFunc 指定一个 Patch 请求
func (m *Module) PatchFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return m.app.router.PatchFunc(path, h)
}
