// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"

	"github.com/issue9/web/dependency"
)

// Module 表示模块信息
type Module struct {
	Name        string
	Deps        []string
	Description string
	Routes      []*Route

	inits []dependency.InitFunc
}

// Route 表示模块信息中的路由信息
type Route struct {
	Path    string
	Handler http.Handler
	Methods []string
}

// Modules 获取当前的所有模块信息
func (app *App) Modules() []*Module {
	return app.modules
}

// AddModule 注册一个新的模块。
func (app *App) AddModule(m *Module) *App {
	app.modules = append(app.modules, m)
	return app
}

func (app *App) initDependency() error {
	dep := dependency.New()

	for _, module := range app.modules {
		dep.Add(module.Name, app.getInit(module), module.Deps...)
	}

	return dep.Init()
}

func (app *App) getInit(m *Module) dependency.InitFunc {
	return func() error {
		for _, init := range m.inits {
			if err := init(); err != nil {
				return err
			}
		}

		for _, r := range m.Routes {
			if err := app.router.Handle(r.Path, r.Handler, r.Methods...); err != nil {
				return err
			}
		}

		return nil
	}
}

// NewModule 声明一个新的模块
func NewModule(name, desc string, deps ...string) *Module {
	return &Module{
		Name:        name,
		Deps:        deps,
		Description: desc,
		Routes:      make([]*Route, 0, 10),
		inits:       make([]dependency.InitFunc, 0, 5),
	}
}

// AddInit 添加一个初始化函数
func (m *Module) AddInit(f dependency.InitFunc) *Module {
	m.inits = append(m.inits, f)
	return m
}

// AddRoute 添加一个路由项
func (m *Module) AddRoute(path string, h http.Handler, methods ...string) *Module {
	m.Routes = append(m.Routes, &Route{
		Methods: methods,
		Path:    path,
		Handler: h,
	})

	return m
}

// Get 指定一个 GET 请求
func (m *Module) Get(path string, h http.Handler) *Module {
	return m.AddRoute(path, h, http.MethodGet)
}

// Post 指定个 POST 请求处理
func (m *Module) Post(path string, h http.Handler) *Module {
	return m.AddRoute(path, h, http.MethodPost)
}

// Delete 指定个 Delete 请求处理
func (m *Module) Delete(path string, h http.Handler) *Module {
	return m.AddRoute(path, h, http.MethodDelete)
}

// Put 指定个 Put 请求处理
func (m *Module) Put(path string, h http.Handler) *Module {
	return m.AddRoute(path, h, http.MethodPut)
}

// Patch 指定个 Patch 请求处理
func (m *Module) Patch(path string, h http.Handler) *Module {
	return m.AddRoute(path, h, http.MethodPatch)
}

// GetFunc 指定一个 GET 请求
func (m *Module) GetFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.AddRoute(path, http.HandlerFunc(h), http.MethodGet)
}

// PostFunc 指定一个 GET 请求
func (m *Module) PostFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.AddRoute(path, http.HandlerFunc(h), http.MethodPost)
}

// DeleteFunc 指定一个 GET 请求
func (m *Module) DeleteFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.AddRoute(path, http.HandlerFunc(h), http.MethodDelete)
}

// PutFunc 指定一个 GET 请求
func (m *Module) PutFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.AddRoute(path, http.HandlerFunc(h), http.MethodPut)
}

// PatchFunc 指定一个 GET 请求
func (m *Module) PatchFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.AddRoute(path, http.HandlerFunc(h), http.MethodPatch)
}
