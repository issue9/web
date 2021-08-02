// SPDX-License-Identifier: MIT

package dep

import (
	"fmt"
	"log"
	"sort"

	"github.com/issue9/sliceutil"
)

// Dep 管理依赖关系
type Dep struct {
	modules []*Module
}

func New() *Dep {
	return &Dep{
		modules: make([]*Module, 0, 50),
	}
}

// Init 触发所有模块下指定名称的函数
func (dep *Dep) Init(l *log.Logger, tag string) error {
	for _, m := range dep.modules { // 检测依赖
		if err := dep.checkDeps(m); err != nil {
			return err
		}
	}

	// 日志不需要标出文件位置。
	flags := l.Flags()
	l.SetFlags(log.Ldate | log.Lmicroseconds)

	l.Printf("开始初始化模块中的 %s...\n", tag)
	for _, m := range dep.modules { // 进行初如化
		if err := dep.initModule(m, l, tag); err != nil {
			return err
		}
	}
	l.Print("初始化完成！\n\n")

	l.SetFlags(flags)

	return nil
}

func (dep *Dep) initModule(m *Module, l *log.Logger, tag string) error {
	for _, depID := range m.deps { // 先初始化依赖项
		depMod := dep.findModule(depID)
		if depMod == nil {
			return fmt.Errorf("模块 %s 依赖项 %s 未找到", m.ID(), depID)
		}

		if err := dep.initModule(depMod, l, tag); err != nil {
			return err
		}
	}

	return m.Init(l, tag)
}

// 检测模块的依赖关系。比如：
// 依赖项是否存在；是否存在自我依赖等。
func (dep *Dep) checkDeps(m *Module) error {
	// 检测依赖项是否都存在
	for _, depID := range m.deps {
		if dep.findModule(depID) == nil {
			return fmt.Errorf("未找到 %s 的依赖模块 %s", m.id, depID)
		}
	}

	if dep.isDep(m.id, m.id) {
		return fmt.Errorf("%s 循环依赖自身", m.id)
	}

	return nil
}

// m1 是否依赖 m2
func (dep *Dep) isDep(m1, m2 string) bool {
	mod1 := dep.findModule(m1)
	if mod1 == nil {
		return false
	}

	for _, depID := range mod1.deps {
		if depID == m2 {
			return true
		}

		if dep.findModule(depID) != nil {
			if dep.isDep(depID, m2) {
				return true
			}
		}
	}

	return false
}

func (dep *Dep) findModule(id string) *Module {
	for _, m := range dep.modules {
		if m.id == id {
			return m
		}
	}
	return nil
}

func (dep *Dep) Tags() []string {
	tags := make([]string, 0, 100)
	for _, m := range dep.modules {
		tags = append(tags, m.Tags()...)
	}
	size := sliceutil.Unique(tags, func(i, j int) bool { return tags[i] == tags[j] })
	tags = tags[:size]
	sort.Strings(tags)
	return tags
}
