// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package modules

import (
	"fmt"
	"sync"

	"github.com/issue9/mux"
	"github.com/issue9/web/module"
)

// 用以表示一个模块所需要的数据。
type mod struct {
	*module.Module
	inited bool // 是否已经初始化
}

// 模块管理工具，管理模块的初始化顺序
type dependency struct {
	modules map[string]*mod
	locker  sync.Mutex
}

// 声明一个 dependency 实例
func newDepencency(ms []*module.Module) *dependency {
	dep := &dependency{
		modules: make(map[string]*mod, len(ms)),
	}

	for _, m := range ms {
		dep.modules[m.Name] = &mod{Module: m}
	}

	return dep
}

// 对所有的模块进行初始化操作，会进行依赖检测。
// 若模块初始化出错，则会中断并返回出错信息。
func (dep *dependency) Init(tag string, router *mux.Prefix) error {
	dep.locker.Lock()
	defer dep.locker.Unlock()

	// 检测依赖
	for _, m := range dep.modules {
		if err := dep.checkDeps(m); err != nil {
			return err
		}
	}

	// 进行初如化
	for _, m := range dep.modules {
		if err := dep.init(m, tag, router); err != nil {
			return err
		}
	}

	return nil
}

// 初始化指定模块，会先初始化其依赖模块。
// 若该模块已经初始化，则不会作任何操作，包括依赖模块的初始化，也不会执行。
func (dep *dependency) init(m *mod, tag string, router *mux.Prefix) error {
	if m.inited {
		return nil
	}

	// 先初始化依赖项
	for _, d := range m.Deps {
		depm, found := dep.modules[d]
		if !found {
			return fmt.Errorf("依赖项[%v]未找到", d)
		}

		if err := dep.init(depm, tag, router); err != nil {
			return err
		}
	}

	// 初始化模块自身
	t := m.Module
	if tag != "" {
		found := false
		if t, found = m.Tags[tag]; !found {
			return nil
		}
	}

	for path, ms := range t.Routes {
		for method, h := range ms {
			if err := router.Handle(path, h, method); err != nil {
				return err
			}
		}
	}

	for _, init := range t.Inits {
		if err := init.F(); err != nil {
			return err
		}
	}

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
