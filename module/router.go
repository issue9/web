// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"net/http"

	"github.com/issue9/mux/v2"
)

// Prefix 声明一个 Prefix 实例。
func (m *Module) Prefix(prefix string) *mux.Prefix {
	return m.ms.router.Prefix(prefix)
}

// Handle 添加一个路由项
func (m *Module) Handle(path string, h http.Handler, methods ...string) *Module {
	if err := m.ms.router.Handle(path, h, methods...); err != nil {
		panic(err)
	}

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
