// SPDX-License-Identifier: MIT

package app

import (
	"net/http"

	"github.com/issue9/mux/v2"
)

// Prefix 声明一个 Prefix 实例。
func (app *App) Prefix(prefix string) *mux.Prefix {
	return app.router.Prefix(prefix)
}

// Handle 添加一个路由项
func (app *App) Handle(path string, h http.Handler, methods ...string) error {
	return app.router.Handle(path, h, methods...)
}

// Get 指定一个 GET 请求
func (app *App) Get(path string, h http.Handler) *mux.Prefix {
	return app.router.Get(path, h)
}

// Post 指定个 POST 请求处理
func (app *App) Post(path string, h http.Handler) *mux.Prefix {
	return app.router.Post(path, h)
}

// Delete 指定个 Delete 请求处理
func (app *App) Delete(path string, h http.Handler) *mux.Prefix {
	return app.router.Delete(path, h)
}

// Put 指定个 Put 请求处理
func (app *App) Put(path string, h http.Handler) *mux.Prefix {
	return app.router.Put(path, h)
}

// Patch 指定个 Patch 请求处理
func (app *App) Patch(path string, h http.Handler) *mux.Prefix {
	return app.router.Patch(path, h)
}

// HandleFunc 指定一个请求
func (app *App) HandleFunc(path string, h func(w http.ResponseWriter, r *http.Request), methods ...string) error {
	return app.router.HandleFunc(path, h, methods...)
}

// GetFunc 指定一个 GET 请求
func (app *App) GetFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return app.router.GetFunc(path, h)
}

// PostFunc 指定一个 Post 请求
func (app *App) PostFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return app.router.PostFunc(path, h)
}

// DeleteFunc 指定一个 Delete 请求
func (app *App) DeleteFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return app.router.DeleteFunc(path, h)
}

// PutFunc 指定一个 Put 请求
func (app *App) PutFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return app.router.PutFunc(path, h)
}

// PatchFunc 指定一个 Patch 请求
func (app *App) PatchFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *mux.Prefix {
	return app.router.PatchFunc(path, h)
}
