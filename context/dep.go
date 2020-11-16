// SPDX-License-Identifier: MIT

package context

import (
	"fmt"
	"log"
)

// 对所有的模块进行初始化操作，会进行依赖检测。
// 若模块初始化出错，则会中断并返回出错信息。
func (srv *Server) initDeps(tag string, info *log.Logger) error {
	// 检测依赖
	for _, m := range srv.modules {
		if err := srv.checkDeps(m); err != nil {
			return err
		}
	}

	// 进行初如化
	for _, m := range srv.modules {
		if err := srv.initModule(m, tag, info); err != nil {
			return err
		}
	}

	return nil
}

// 初始化指定模块，会先初始化其依赖模块。
//
// 若该模块已经初始化，则不会作任何操作，包括依赖模块的初始化，也不会执行。
// 若 tag 不为空，表示只调用该标签下的初始化函数。
func (srv *Server) initModule(m *Module, tag string, info *log.Logger) error {
	if m.inited {
		return nil
	}

	// 先初始化依赖项
	for _, d := range m.Deps {
		depm := srv.module(d)
		if depm == nil {
			return fmt.Errorf("依赖项[%s]未找到", d)
		}

		if err := srv.initModule(depm, tag, info); err != nil {
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

	info.Println("开始初始化模块：", m.Name)

	// 执行当前模块的初始化函数
	for _, init := range inits {
		title := init.title

		info.Println("  执行初始化函数：", title)
		if err := init.f(); err != nil {
			return err
		}
	} // end for

	m.inited = true
	return nil
}

// 检测模块的依赖关系。比如：
// 依赖项是否存在；是否存在自我依赖等。
func (srv *Server) checkDeps(m *Module) error {
	// 检测依赖项是否都存在
	for _, d := range m.Deps {
		if srv.module(d) == nil {
			return fmt.Errorf("未找到[%v]的依赖模块[%v]", m.Name, d)
		}
	}

	if srv.isDep(m.Name, m.Name) {
		return fmt.Errorf("存在循环依赖项:[%v]", m.Name)
	}

	return nil
}

// m1 是否依赖 m2
func (srv *Server) isDep(m1, m2 string) bool {
	module1 := srv.module(m1)
	if module1 == nil {
		return false
	}

	for _, d := range module1.Deps {
		if d == m2 {
			return true
		}

		if srv.module(d) != nil {
			if srv.isDep(d, m2) {
				return true
			}
		}
	}

	return false
}

func (srv *Server) module(name string) *Module {
	for _, m := range srv.modules {
		if m.Name == name {
			return m
		}
	}
	return nil
}
