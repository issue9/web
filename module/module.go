// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package module 提供模块的的相关功能。
package module

import (
	"net/http"

	"github.com/issue9/mux"
	"github.com/issue9/web/internal/dependency"
)

// Type 表示模块的类型
type Type int8

// 表示模块的类型
const (
	TypeModule Type = iota + 1
	TypePlugin
)

// Module 表示模块信息
type Module struct {
	Type        Type
	Name        string
	Deps        []string
	Description string

	// 当前模块的所有路由项。
	// 键名为路由地址，键值为路由中启用的请求方法。
	Routes map[string][]string

	// 当前模块的安装功能。
	//
	// 键名指定了安装的版本，键值则为安装脚本。
	installs map[string][]*task

	inits  []func() error
	router *mux.Prefix
}

// New 声明一个新的模块
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func New(router *mux.Prefix, name, desc string, deps ...string) *Module {
	return &Module{
		Type:        TypeModule,
		Name:        name,
		Deps:        deps,
		Description: desc,
		Routes:      make(map[string][]string, 10),
		inits:       make([]func() error, 0, 5),
		installs:    make(map[string][]*task, 10),
		router:      router,
	}
}

// GetInit 将 Module 的内容生成一个 dependency.InitFunc 函数
func (m *Module) GetInit() dependency.InitFunc {
	return func() error {
		for _, init := range m.inits {
			if err := init(); err != nil {
				return err
			}
		}

		return nil
	}
}

// Mux 返回 github.com/issue9/mxu.Mux 实例
func (m *Module) Mux() *mux.Mux {
	return m.router.Mux()
}

// AddInit 添加一个初始化函数
func (m *Module) AddInit(f func() error) *Module {
	m.inits = append(m.inits, f)
	return m
}

// Handle 添加一个路由项
func (m *Module) Handle(path string, h http.Handler, methods ...string) *Module {
	if err := m.router.Handle(path, h, methods...); err != nil {
		panic(err)
	}

	route, found := m.Routes[path]
	if !found {
		route = make([]string, 0, 10)
	}
	m.Routes[path] = append(route, methods...)

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
