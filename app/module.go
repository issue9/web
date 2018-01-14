// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"

	"github.com/issue9/mux"
	"github.com/issue9/web/dependency"
)

// InitFunc 模块的初始化函数
type InitFunc func(r *mux.Prefix) error

// Module 表示模块信息
type Module struct {
	Name        string
	Deps        []string
	Init        dependency.InitFunc
	Description string
	Routes      []*Route
}

// Route 表示模块信息中的路由信息
type Route struct {
	Method  string
	Path    string
	Handler http.Handler
}

func (app *App) initDependency() error {
	dep := dependency.New()

	for _, module := range app.modules {
		dep.AddModule(&dependency.Module{
			Name: module.Name,
			Deps: module.Deps,
			Init: app.getInit(module),
		})
	}

	return dep.Init()
}

func (app *App) getInit(m *Module) dependency.InitFunc {
	return func() error {
		if m.Init != nil {
			if err := m.Init(); err != nil {
				return err
			}
		}

		for _, r := range m.Routes {
			if err := app.Router().Handle(r.Path, r.Handler, r.Method); err != nil {
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
	}
}

// AddRoute 添加一个路由项
func (m *Module) AddRoute(method, path string, h http.Handler) *Module {
	m.Routes = append(m.Routes, &Route{
		Method:  method,
		Path:    path,
		Handler: h,
	})

	return m
}
