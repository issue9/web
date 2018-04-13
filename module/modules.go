// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"

	"github.com/issue9/mux"
	"github.com/issue9/web/internal/dependency"
)

// ErrModulesIsInited 当前模块已经被初始化。不能再次执行初始化操作
var ErrModulesIsInited = errors.New("当前的所有模块已经初始化")

// Modules 模块的管理功能
type Modules struct {
	router  *mux.Prefix
	modules []*Module
	inited  bool
}

// NewModules 新建一个模块管理功能
//
// router 指定所有模块中的路由将被添加到该路径之下。
func NewModules(router *mux.Prefix) *Modules {
	if router == nil {
		panic("router 不能为空")
	}

	return &Modules{
		router:  router,
		modules: make([]*Module, 0, 10),
	}
}

// Modules 返回所有的模块信息
func (ms *Modules) Modules() []*Module {
	return ms.modules
}

// Init 执行初始化操作
func (ms *Modules) Init() error {
	if ms.inited {
		return ErrModulesIsInited
	}

	dep := dependency.New()

	for _, module := range ms.modules {
		if err := dep.Add(module.Name, ms.getInit(module), module.Deps...); err != nil {
			return err
		}
	}

	if err := dep.Init(); err != nil {
		return err
	}

	ms.inited = true
	return nil
}

// 将 Module 的内容生成一个 dependency.InitFunc 函数
func (ms *Modules) getInit(m *Module) dependency.InitFunc {
	return func() error {
		for _, init := range m.inits {
			if err := init(); err != nil {
				return err
			}
		}

		for _, r := range m.Routes {
			for method, handler := range r.Methods {
				if m.middleware != nil {
					handler = m.middleware(handler)
				}

				if err := ms.router.Handle(r.Path, handler, method); err != nil {
					return err
				}
			}
		}

		return nil
	}
}
