// SPDX-License-Identifier: MIT

// Package dep 依赖管理
package dep

import (
	"fmt"
	"log"
)

// Init 对所有的模块进行初始化操作
//
// 会进行依赖检测。若模块初始化出错，则会中断并返回出错信息。
func Init(ms []Module, info *log.Logger) error {
	// 检测依赖
	for _, m := range ms {
		if err := checkDeps(ms, m); err != nil {
			return err
		}
	}

	// 进行初如化
	for _, m := range ms {
		if err := initModule(ms, m, info); err != nil {
			return err
		}
	}

	return nil
}

// 初始化指定模块，会先初始化其依赖模块。
//
// 若该模块已经初始化，则不会作任何操作，包括依赖模块的初始化，也不会执行。
// 若 tag 不为空，表示只调用该标签下的初始化函数。
func initModule(ms []Module, m Module, info *log.Logger) error {
	if m.Inited() {
		return nil
	}

	// 先初始化依赖项
	for _, d := range m.Deps() {
		depm := findModule(ms, d)
		if depm == nil {
			return fmt.Errorf("依赖项 %s 未找到", d)
		}

		if err := initModule(ms, depm, info); err != nil {
			return err
		}
	}

	info.Println("开始初始化模块：", m.ID())

	return m.Init()
}

// 检测模块的依赖关系。比如：
// 依赖项是否存在；是否存在自我依赖等。
func checkDeps(ms []Module, m Module) error {
	// 检测依赖项是否都存在
	for _, d := range m.Deps() {
		if findModule(ms, d) == nil {
			return fmt.Errorf("未找到 %s 的依赖模块 %s", m.ID(), d)
		}
	}

	if isDep(ms, m.ID(), m.ID()) {
		return fmt.Errorf("存在循环依赖项:%s", m.ID())
	}

	return nil
}

// m1 是否依赖 m2
func isDep(ms []Module, m1, m2 string) bool {
	module1 := findModule(ms, m1)
	if module1 == nil {
		return false
	}

	for _, d := range module1.Deps() {
		if d == m2 {
			return true
		}

		if findModule(ms, d) != nil {
			if isDep(ms, d, m2) {
				return true
			}
		}
	}

	return false
}

func findModule(ms []Module, id string) Module {
	for _, m := range ms {
		if m.ID() == id {
			return m
		}
	}
	return nil
}
