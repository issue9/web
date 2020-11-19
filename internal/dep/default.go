// SPDX-License-Identifier: MIT

package dep

import (
	"fmt"
	"log"
)

// Initialization 表示模块初始化函数的基本数据
type Initialization struct {
	title string
	f     func() error
}

// Default 默认的 Module 实现
//
// 包含了一个函数列表，当作模块的初始化功能。
type Default struct {
	id     string
	desc   string
	deps   []string
	inited bool
	inits  []Initialization
}

// NewDefaultModule 返回 Default 实例
func NewDefaultModule(id, desc string, dep ...string) *Default {
	return &Default{
		id:    id,
		desc:  desc,
		deps:  dep,
		inits: make([]Initialization, 0, 5),
	}
}

// ID 唯一 ID
func (m *Default) ID() string {
	return m.id
}

// Description 详细描述
func (m *Default) Description() string {
	return m.desc
}

// Deps 返回依赖列表
func (m *Default) Deps() []string {
	return m.deps
}

// Inited 是否已经初始化
func (m *Default) Inited() bool {
	return m.inited
}

// Init 执行初始化函数
//
// 如果包含了多个函数，在其中一个函数出错之后，将退出执行列表。
func (m *Default) Init(info *log.Logger) error {
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
func (m *Default) AddInit(f func() error, title string) {
	if m.Inited() {
		panic(fmt.Sprintf("模块 %s 已经初始化，不能再添加初始化函数", m.ID()))
	}

	m.inits = append(m.inits, Initialization{title: title, f: f})
}
