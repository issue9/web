// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package module 提供模块的的相关功能。
package module

import (
	"fmt"
	"net/http"
)

// Type 表示模块的类型
type Type int8

// 表示模块的类型
const (
	TypeModule Type = iota + 1
	TypePlugin
)

var defaultMethods = []string{
	http.MethodDelete,
	http.MethodGet,
	http.MethodOptions,
	http.MethodPatch,
	http.MethodPost,
	http.MethodPut,
	http.MethodTrace,
	http.MethodConnect,
}

// Module 表示模块信息
type Module struct {
	Type        Type
	Name        string
	Deps        []string
	Description string

	// 第一个键名为路径，第二键名为请求方法
	Routes map[string]map[string]http.Handler

	// 一些初始化函数
	Inits []*Init

	Tags map[string]*Module
}

// Init 表示初始化功能的相关数据
type Init struct {
	Title string
	F     func() error
}

// New 声明一个新的模块
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func New(name, desc string, deps ...string) *Module {
	return &Module{
		Type:        TypeModule,
		Name:        name,
		Deps:        deps,
		Description: desc,
		Routes:      make(map[string]map[string]http.Handler, 10),
		Inits:       make([]*Init, 0, 5),
		Tags:        make(map[string]*Module, 5),
	}
}

// NewTag 为当前模块添加某一版本号下的安装脚本。
func (m *Module) NewTag(tag string) *Module {
	if _, found := m.Tags[tag]; !found {
		m.Tags[tag] = &Module{
			Type:        TypeModule,
			Name:        "",
			Deps:        nil,
			Description: "",
			Routes:      make(map[string]map[string]http.Handler, 10),
			Inits:       make([]*Init, 0, 5),
		}
	}

	return m.Tags[tag]
}

// AddInit 添加一个初始化函数
func (m *Module) AddInit(f func() error) *Module {
	m.Inits = append(m.Inits, &Init{F: f})
	return m
}

// Task 添加一条安装脚本
func (m *Module) Task(title string, fn func() error) {
	m.Inits = append(m.Inits, &Init{Title: title, F: fn})
}

// Handle 添加一个路由项
func (m *Module) Handle(path string, h http.Handler, methods ...string) *Module {
	ms, found := m.Routes[path]
	if !found {
		ms = make(map[string]http.Handler, 8)
		m.Routes[path] = ms
	}

	if len(methods) == 0 {
		methods = defaultMethods
	}
	for _, method := range methods {
		if _, found = ms[method]; found {
			panic(fmt.Sprintf("路径 %s 已经存在相同的请求方法 %s", path, method))
		}
		ms[method] = h
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
