// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package dependency 依赖管理，可以保证各个模块能按顺序初始化依赖项。
package dependency

import (
	"fmt"
	"sync"
)

// InitFunc 模块的初始化函数。
type InitFunc func() error

// Module 用以表示一个模块所需要的数据。
type Module struct {
	Name   string   // 模块名
	Init   InitFunc // 初始化函数
	Deps   []string // 依赖项
	inited bool     // 是否已经初始化
}

// Dependency 模块管理工具，管理模块的初始化顺序
type Dependency struct {
	modules map[string]*Module
	lock    sync.Mutex
}

// New 声明一个 Dependency 实例
func New() *Dependency {
	return &Dependency{
		modules: make(map[string]*Module, 20),
	}
}

// Add 添加一个新的模块。
//
// name 为模块的名称；
// init 为模块的初始化函数；
// deps 为模块的依赖模块，依赖模块可以后于当前模块注册，但必须要存在。
func (dep *Dependency) Add(name string, init InitFunc, deps ...string) error {
	return dep.AddModule(&Module{
		Name: name,
		Init: init,
		Deps: deps,
	})
}

// AddModule 添加一个新的模块信息
func (dep *Dependency) AddModule(m *Module) error {
	dep.lock.Lock()
	defer dep.lock.Unlock()

	if _, found := dep.modules[m.Name]; found {
		return fmt.Errorf("模块[%v]已经存在", m.Name)
	}

	dep.modules[m.Name] = m
	return nil
}

// Init 对所有的模块进行初始化操作，会进行依赖检测。
// 若模块初始化出错，则会中断并返回出错信息。
func (dep *Dependency) Init() error {
	dep.lock.Lock()
	defer dep.lock.Unlock()

	for _, m := range dep.modules {
		if err := dep.checkDeps(m); err != nil {
			return err
		}
	}

	for _, m := range dep.modules {
		if err := dep.init(m); err != nil {
			return err
		}
	}

	return nil
}

// 初始化指定模块，会先初始化其依赖模块。
// 若该模块已经初始化，则不会作任何操作，包括依赖模块的初始化，也不会执行。
func (dep *Dependency) init(m *Module) error {
	if m.inited {
		return nil
	}

	// 先初始化依赖项
	for _, d := range m.Deps {
		depm, found := dep.modules[d]
		if !found {
			return fmt.Errorf("依赖项[%v]未找到", d)
		}

		if err := dep.init(depm); err != nil {
			return err
		}
	}

	// 初始化模块自身
	if err := m.Init(); err != nil {
		return err
	}

	m.inited = true
	return nil
}

// 检测模块的依赖关系。比如：
// 依赖项是否存在；是否存在自我依赖等。
func (dep *Dependency) checkDeps(m *Module) error {
	// 检测依赖项是否都存在
	for _, d := range m.Deps {
		_, found := dep.modules[d]
		if !found {
			return fmt.Errorf("未找到[%v]的依赖模块[%v]", m.Name, d)
		}
	}

	if dep.isDep(m.Name, m.Name) {
		return fmt.Errorf("存在循环依赖项:[%v]", m.Name)
	}

	return nil
}

// m1 是否依赖 m2
func (dep *Dependency) isDep(m1, m2 string) bool {
	module1, found := dep.modules[m1]
	if !found {
		return false
	}

	for _, d := range module1.Deps {
		if d == m2 {
			return true
		}

		if _, found = dep.modules[d]; found {
			if dep.isDep(d, m2) {
				return true
			}
		}
	}

	return false
}
