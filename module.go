// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/issue9/mux"
)

var (
	serveMux  = mux.NewServeMux()
	modules   = map[string]*Module{} // 所有模块的列表。
	modulesMu sync.Mutex
)

var ErrModuleExists = errors.New("该名称的模块已经存在")

// 模块化管理路由项。相对于mux.Group，添加了模块依赖管理。
type Module struct {
	Name         string   // 名称
	Dependencies []string // 依赖项
	group        *mux.Group
}

// 所有模块列表。
func Modules() []*Module {
	ret := make([]*Module, 0, len(modules))
	for _, m := range modules {
		ret = append(ret, m)
	}
	return ret
}

// 获取指定名称的模块，若不存在，则返回nil
func GetModule(name string) *Module {
	modulesMu.Lock()
	defer modulesMu.Unlock()

	return modules[name]
}

// 声明一个新的模块，若该名称已经存在，则返回错误信息。
// name 模块名称
// dependencies 该模块的依赖模块列表。
func NewModule(name string, dependencies ...string) (*Module, error) {
	modulesMu.Lock()
	defer modulesMu.Unlock()

	// 确保没有同名存在。
	if _, found := modules[name]; found {
		return nil, ErrModuleExists
	}

	// 检测依赖模块是否都已经存在
	for _, m := range dependencies {
		if _, found := modules[m]; !found {
			return nil, fmt.Errorf("依赖项[%v]不存在", m)
		}
	}

	m := &Module{
		Name:         name,
		Dependencies: dependencies,
		group:        serveMux.Group(),
	}
	modules[name] = m

	return m, nil
}

// 当前模块的路由是否处于运行状态
func (m *Module) IsRunning() bool {
	return m.group.IsRunning()
}

// 将当前模块改为运行状态
func (m *Module) Start() {
	m.group.Start()
}

// 将当前模块改为暂停状态。
func (m *Module) Stop() {
	m.group.Stop()
}

// 添加一个路由项。
// 具体参数说明，可参考github.com/issue9/mux.ServeMux.Add()方法。
func (m *Module) Add(pattern string, h http.Handler, methods ...string) *Module {
	m.group.Add(pattern, h, methods...)
	return m
}

// Get 相当于Module.Add(pattern, h, "GET")
func (m *Module) Get(pattern string, h http.Handler) *Module {
	return m.Add(pattern, h, "GET")
}

// Post 相当于Module.Add(pattern, h, "POST")
func (m *Module) Post(pattern string, h http.Handler) *Module {
	return m.Add(pattern, h, "POST")
}

// Delete 相当于Module.Add(pattern, h, "DELETE")
func (m *Module) Delete(pattern string, h http.Handler) *Module {
	return m.Add(pattern, h, "DELETE")
}

// Put 相当于Module.Add(pattern, h, "PUT")
func (m *Module) Put(pattern string, h http.Handler) *Module {
	return m.Add(pattern, h, "PUT")
}

// Patch 相当于Module.Add(pattern, h, "PATCH")
func (m *Module) Patch(pattern string, h http.Handler) *Module {
	return m.Add(pattern, h, "PATCH")
}

// Any 相当于Module.Add(pattern, h)
func (m *Module) Any(pattern string, h http.Handler) *Module {
	m.group.Any(pattern, h)
	return m
}

// AddFunc 相当于Module.Add(pattern, http.HandlerFunc(fun), methods...)
func (m *Module) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *Module {
	m.group.AddFunc(pattern, fun, methods...)
	return m
}

// GetFunc 相当于Module.AddFunc(pattern, fun, "GET")
func (m *Module) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Module {
	return m.AddFunc(pattern, fun, "GET")
}

// PutFunc 相当于Module.AddFunc(pattern, fun, "PUT")
func (m *Module) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Module {
	return m.AddFunc(pattern, fun, "PUT")
}

// PostFunc 相当于Module.AddFunc(pattern, fun, "POST")
func (m *Module) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Module {
	return m.AddFunc(pattern, fun, "POST")
}

// DeleteFunc 相当于Module.Addunc(pattern, fun, "DELETE")
func (m *Module) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Module {
	return m.AddFunc(pattern, fun, "DELETE")
}

// PatchFunc 相当于Module.AddFunc(pattern, fun, "PATCH")
func (m *Module) PatchFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Module {
	return m.AddFunc(pattern, fun, "PATCH")
}

// AnyFunc 相当于Module.AddFunc(pattern, fn)
func (m *Module) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Module {
	m.group.AnyFunc(pattern, fun)
	return m
}

// 创建一个mux.Prefix对象，具体可参考该实例说明。
func (m *Module) Prefix(prefix string) *mux.Prefix {
	return m.group.Prefix(prefix)
}
