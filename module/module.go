// SPDX-License-Identifier: MIT

// Package module 模块之间的依赖管理
package module

import (
	"fmt"

	"github.com/issue9/logs/v2"
)

// Module 模块
type Module struct {
	*initializer
	desc   string
	deps   []string
	inited bool

	// 与特定标签关联的初始化函数，这些函数默认情况下不会被调用。
	tags map[string]*initializer
}

// NewModule 返回 Default 实例
func NewModule(name, desc string, dep ...string) *Module {
	return &Module{
		initializer: &initializer{name: name, inits: make([]*initializer, 0, 5)},
		desc:        desc,
		deps:        dep,
	}
}

// ID 模块的唯一名称
func (m *Module) ID() string { return m.name }

// Description 模块的详细描述信息
func (m *Module) Description() string { return m.desc }

// Deps 模块的依赖模块
func (m *Module) Deps() (deps []string) {
	if l := len(m.deps); l > 0 {
		deps = make([]string, l)
		copy(deps, m.deps)
	}
	return
}

// Inited 是否已经初始化
func (m *Module) Inited() bool { return m.inited }

// Init 初始化当前模块
//
// t 如果不为空，则表示只调用附加于该标签的初始化函数，且不会考虑是否已经初始化；
// l 表示初始化信息的调用情况输出通道；
func (m *Module) Init(t string, l *logs.Logs) error {
	if t != "" {
		if i := m.tags[t]; i != nil {
			return i.init(l, 0)
		}
		return nil
	}

	if m.Inited() {
		err := fmt.Errorf("模块 %s 已经初始化", m.ID())
		l.Error(err)
		return err
	}

	if err := m.init(l, 0); err != nil {
		return err
	}
	m.inited = true
	return nil
}

// AddInit 实现 Initializer 接口
func (m *Module) AddInit(title string, f func() error) Initializer {
	return m.initializer.AddInit(title, f)
}

// GetTag 获取特定名称的初始化函数
func (m *Module) GetTag(tag string) Initializer {
	if m.tags == nil {
		m.tags = map[string]*initializer{}
	}

	if _, found := m.tags[tag]; !found {
		m.tags[tag] = &initializer{name: tag}
	}
	return m.tags[tag]
}
