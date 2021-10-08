// SPDX-License-Identifier: MIT

// Package dep 依赖关系管理
package dep

import (
	"log"

	"github.com/issue9/localeutil"
)

// Dep 初始化依赖项
func Dep(info *log.Logger, items []*Item) error {
	for _, m := range items { // 检测依赖
		if err := checkDeps(items, m); err != nil {
			return err
		}
	}

	// 日志不需要标出文件位置。
	flags := info.Flags()
	info.SetFlags(log.Ldate | log.Lmicroseconds)

	for _, m := range items { // 进行初如化
		if err := initItem(items, m, info); err != nil {
			return err
		}
	}

	info.SetFlags(flags)

	return nil
}

func initItem(items []*Item, m *Item, l *log.Logger) error {
	for _, depID := range m.deps { // 先初始化依赖项
		depMod := findItem(items, depID)
		if depMod == nil {
			return localeutil.Error("not found dependence", m.id, depID)
		}

		if err := initItem(items, depMod, l); err != nil {
			return err
		}
	}

	if m.called {
		return nil
	}

	l.Println(m.id, "...")

	err := m.call(l)
	if err != nil {
		l.Printf("%s [FAIL:%s]\n\n", m.id, err.Error())
	} else {
		l.Printf("%s [OK]\n\n", m.id)
	}
	return err
}

// 检测模块的依赖关系。比如：
// 依赖项是否存在；是否存在自我依赖等。
func checkDeps(items []*Item, m *Item) error {
	// 检测依赖项是否都存在
	for _, depID := range m.deps {
		if findItem(items, depID) == nil {
			return localeutil.Error("not found dependence", m.id, depID)
		}
	}

	if IsDep(items, m.id, m.id) {
		return localeutil.Error("cyclic dependence", m.id)
	}

	return nil
}

// IsDep m1 是否依赖 m2
func IsDep(items []*Item, m1, m2 string) bool {
	mod1 := findItem(items, m1)
	if mod1 == nil {
		return false
	}

	for _, depID := range mod1.deps {
		if depID == m2 {
			return true
		}

		if findItem(items, depID) != nil {
			if IsDep(items, depID, m2) {
				return true
			}
		}
	}

	return false
}

func findItem(items []*Item, id string) *Item {
	for _, m := range items {
		if m.id == id {
			return m
		}
	}
	return nil
}
