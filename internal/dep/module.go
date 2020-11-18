// SPDX-License-Identifier: MIT

package dep

// Module 需要处理依赖关系的模块需要实现的接口
type Module interface {
	// 模块的唯一
	ID() string

	// 当前模块依赖的其它模块 ID
	Deps() []string

	// 对该模块进行初始化操作
	Init() error

	// 当前模块是否已经被初始化
	Inited() bool
}
