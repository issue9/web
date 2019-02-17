// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package module 提供模块的的相关功能。
package module

import "net/http"

// Type 表示模块的类型
type Type int8

// 表示模块的类型
const (
	TypeModule Type = iota + 1
	TypePlugin
	TypeTag
)

// Module 表示模块信息
type Module struct {
	Type        Type
	Name        string
	Description string
	Deps        []string
	Tags        map[string]*Module
	Inits       []*Init
	Services    []*Service

	// 路由项列表。
	//
	// 第一个键名为路径，第二键名为请求方法
	Routes map[string]map[string]http.Handler
}

// New 声明一个新的模块
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
//
// 仅供框架内部使用，不保证函数签名的兼容性。
func New(typ Type, name, desc string, deps ...string) *Module {
	return &Module{
		Type:        typ,
		Name:        name,
		Description: desc,
		Deps:        deps,
		Routes:      make(map[string]map[string]http.Handler, 10),
	}
}

// NewTag 为当前模块生成特定名称的子模块。
// 若已经存在，则直接返回该子模块。
func (m *Module) NewTag(tag string) *Module {
	if m.Type == TypeTag {
		panic("子模块不能再添子模块")
	}

	if m.Tags == nil {
		m.Tags = make(map[string]*Module, 5)
	}

	if _, found := m.Tags[tag]; !found {
		m.Tags[tag] = New(TypeTag, tag, "")
	}

	return m.Tags[tag]
}
