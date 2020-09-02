// SPDX-License-Identifier: MIT

package module

import (
	"net/http"

	"github.com/issue9/mux/v2"
)

// Prefix 声明一个 Prefix 实例。
func (m *Module) Prefix(prefix string) *mux.Prefix {
	return m.ms.app.Prefix(prefix)
}

// Handle 添加一个路由项
func (m *Module) Handle(path string, h http.Handler, methods ...string) error {
	return m.ms.app.Handle(path, h, methods...)
}

// Get 指定一个 GET 请求
func (m *Module) Get(path string, h http.Handler) *mux.Prefix {
	return m.ms.app.Get(path, h)
}

// Post 指定个 POST 请求处理
func (m *Module) Post(path string, h http.Handler) *mux.Prefix {
	return m.ms.app.Post(path, h)
}

// Delete 指定个 Delete 请求处理
func (m *Module) Delete(path string, h http.Handler) *mux.Prefix {
	return m.ms.app.Delete(path, h)
}

// Put 指定个 Put 请求处理
func (m *Module) Put(path string, h http.Handler) *mux.Prefix {
	return m.ms.app.Put(path, h)
}

// Patch 指定个 Patch 请求处理
func (m *Module) Patch(path string, h http.Handler) *mux.Prefix {
	return m.ms.app.Patch(path, h)
}

// HandleFunc 指定一个请求
func (m *Module) HandleFunc(path string, h func(w http.ResponseWriter, r *http.Request), methods ...string) error {
	return m.ms.app.HandleFunc(path, h, methods...)
}

// GetFunc 指定一个 GET 请求
func (m *Module) GetFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return m.ms.app.GetFunc(path, h)
}

// PostFunc 指定一个 Post 请求
func (m *Module) PostFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return m.ms.app.PostFunc(path, h)
}

// DeleteFunc 指定一个 Delete 请求
func (m *Module) DeleteFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return m.ms.app.DeleteFunc(path, h)
}

// PutFunc 指定一个 Put 请求
func (m *Module) PutFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return m.ms.app.PutFunc(path, h)
}

// PatchFunc 指定一个 Patch 请求
func (m *Module) PatchFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return m.ms.app.PatchFunc(path, h)
}
