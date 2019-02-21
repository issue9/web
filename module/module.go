// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package module 提供模块的的相关功能。
package module

// Module 表示模块信息
type Module struct {
	Name        string
	Description string
	Deps        []string

	tags  map[string]*Module
	inits []*initialization
	ms    *Modules

	inited bool
}

// 声明一个新的模块
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
//
// 仅供框架内部使用，不保证函数签名的兼容性。
func newModule(ms *Modules, name, desc string, deps ...string) *Module {
	return &Module{
		Name:        name,
		Description: desc,
		Deps:        deps,
		ms:          ms,
	}
}

// NewTag 为当前模块生成特定名称的子模块。
// 若已经存在，则直接返回该子模块。
func (m *Module) NewTag(tag string) *Module {
	if m.tags == nil {
		m.tags = make(map[string]*Module, 5)
	}

	if _, found := m.tags[tag]; !found {
		m.tags[tag] = newModule(m.ms, tag, "")
	}

	return m.tags[tag]
}

// NewModule 声明一个新的模块
func (ms *Modules) NewModule(name, desc string, deps ...string) *Module {
	m := newModule(ms, name, desc, deps...)
	ms.appendModules(m)
	return m
}
