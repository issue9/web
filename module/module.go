// SPDX-License-Identifier: MIT

// Package module 提供模块管理的相关功能
package module

import (
	"sort"
)

// InstallFunc 安装模块的函数签名
type InstallFunc func(*Server)

// Module 表示模块信息
//
// 模块仅作为在初始化时在代码上的一种分类，一旦初始化完成，
// 则不再有模块的概念，修改模块的相关属性，也不会对代码有实质性的改变。
type Module struct {
	Tag
	Name        string
	Description string
	Deps        []string
	tags        map[string]*Tag
	srv         *Server
}

// Tag 表示与特定标签相关联的初始化函数列表
//
// 依附于模块，共享模块的依赖关系。
//
// 一般是各个模块下的安装脚本使用。
type Tag struct {
	inits []*initialization
}

// 表示初始化功能的相关数据
type initialization struct {
	title string
	f     func() error
}

// NewTag 为当前模块生成特定名称的子模块
//
// 若已经存在，则直接返回该子模块。
//
// Tag 是依赖关系与当前模块相同，但是功能完全独立的模块，
// 一般用于功能更新等操作。
func (m *Module) NewTag(tag string) *Tag {
	if m.tags == nil {
		m.tags = make(map[string]*Tag, 5)
	}

	if _, found := m.tags[tag]; !found {
		m.tags[tag] = &Tag{
			inits: make([]*initialization, 0, 5),
		}
	}

	return m.tags[tag]
}

// NewModule 声明一个新的模块
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func (srv *Server) NewModule(name, desc string, deps ...string) *Module {
	m := &Module{
		Name:        name,
		Description: desc,
		Deps:        deps,
		srv:         srv,
	}
	srv.modules = append(srv.modules, m)
	return m
}

// AddInit 添加一个初始化函数
//
// title 该初始化函数的名称。
func (t *Tag) AddInit(f func() error, title string) *Tag {
	if t.inits == nil {
		t.inits = make([]*initialization, 0, 5)
	}

	t.inits = append(t.inits, &initialization{f: f, title: title})
	return t
}

// Tags 返回所有的子模块名称
//
// 键名为模块名称，键值为该模块下的标签列表。
func (srv *Server) Tags() map[string][]string {
	ret := make(map[string][]string, len(srv.modules)*2)

	for _, m := range srv.modules {
		tags := make([]string, 0, len(m.tags))
		for k := range m.tags {
			tags = append(tags, k)
		}
		sort.Strings(tags)
		ret[m.Name] = tags
	}

	return ret
}

// Modules 当前系统使用的所有模块信息
func (srv *Server) Modules() []*Module {
	return srv.modules
}
