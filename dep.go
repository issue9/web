// SPDX-License-Identifier: MIT

package web

import (
	"fmt"
	"log"
)

type mod struct {
	*Module
	inited bool
}

// 模块管理工具，管理模块的初始化顺序
type dependency struct {
	modules map[string]*mod
	l       *log.Logger
}

// l 表示输出一些执行过程中的提示信息
func newDepencency(ms []*Module, l *log.Logger) *dependency {
	dep := &dependency{
		modules: make(map[string]*mod, len(ms)),
		l:       l,
	}

	for _, m := range ms {
		dep.modules[m.Name] = &mod{Module: m}
	}

	return dep
}

// 对所有的模块进行初始化操作，会进行依赖检测。
// 若模块初始化出错，则会中断并返回出错信息。
func (dep *dependency) init(tag string) error {
	// 检测依赖
	for _, m := range dep.modules {
		if err := dep.checkDeps(m); err != nil {
			return err
		}
	}

	// 进行初如化
	for _, m := range dep.modules {
		if err := dep.initModule(m, tag); err != nil {
			return err
		}
	}

	return nil
}

// 初始化指定模块，会先初始化其依赖模块。
//
// 若该模块已经初始化，则不会作任何操作，包括依赖模块的初始化，也不会执行。
// 若 tag 不为空，表示只调用该标签下的初始化函数。
func (dep *dependency) initModule(m *mod, tag string) error {
	if m.inited {
		return nil
	}

	// 先初始化依赖项
	for _, d := range m.Deps {
		depm, found := dep.modules[d]
		if !found {
			return fmt.Errorf("依赖项[%s]未找到", d)
		}

		if err := dep.initModule(depm, tag); err != nil {
			return err
		}
	}

	inits := m.inits
	if tag != "" {
		t, found := m.tags[tag]
		if !found {
			return nil
		}
		inits = t.inits
	}

	dep.l.Println("开始初始化模块：", m.Name)

	// 执行当前模块的初始化函数
	for _, init := range inits {
		title := init.title

		dep.l.Println("  执行初始化函数：", title)
		if err := init.f(); err != nil {
			return err
		}
	} // end for

	m.inited = true
	return nil
}

// 检测模块的依赖关系。比如：
// 依赖项是否存在；是否存在自我依赖等。
func (dep *dependency) checkDeps(m *mod) error {
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
func (dep *dependency) isDep(m1, m2 string) bool {
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