// SPDX-License-Identifier: MIT

package dep

import (
	"fmt"
	"log"
)

// 表示模块初始化函数的基本数据
type initialization struct {
	title string
	f     func() error
}

// Module 默认的 Module 实现
//
// 包含了一个函数列表，当作模块的初始化功能。
type Module struct {
	id     string
	desc   string
	deps   []string
	inited bool
	inits  []initialization
	items  map[string]*Module
}

// NewModule 返回 Default 实例
func NewModule(id, desc string, dep ...string) *Module {
	return &Module{
		id:    id,
		desc:  desc,
		deps:  dep,
		inits: make([]initialization, 0, 5),
	}
}

// ID 唯一 ID
func (m *Module) ID() string {
	return m.id
}

// Description 详细描述
func (m *Module) Description() string {
	return m.desc
}

// Deps 返回依赖列表
func (m *Module) Deps() []string {
	return m.deps
}

// Inited 是否已经初始化
func (m *Module) Inited() bool {
	return m.inited
}

// Init 执行初始化函数
//
// 如果包含了多个函数，在其中一个函数出错之后，将退出执行列表。
func (m *Module) Init(info *log.Logger) error {
	for _, init := range m.inits {
		info.Printf("执行初始化函数：%s\n", init.title)
		if err := init.f(); err != nil {
			return err
		}
	}

	m.inited = true
	return nil
}

// AddInit 添加初始化函数
func (m *Module) AddInit(title string, f func() error) {
	if m.Inited() {
		panic(fmt.Sprintf("模块 %s 已经初始化，不能再添加初始化函数", m.ID()))
	}

	m.inits = append(m.inits, initialization{title: title, f: f})
}

// New 获取指定名称的子模块
func (m *Module) New(name string) *Module {
	if m.Inited() {
		panic(fmt.Sprintf("模块 %s 已经初始化，不能再添加初始化函数", m.ID()))
	}

	if m.items == nil {
		m.items = make(map[string]*Module, 2)
	}

	if item, found := m.items[name]; found {
		return item
	}
	item := NewModule(m.id, m.desc, m.deps...)
	m.items[name] = item
	return item
}
