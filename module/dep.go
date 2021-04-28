// SPDX-License-Identifier: MIT

package module

import (
	"fmt"
	"sort"

	"github.com/issue9/logs/v2"
	"github.com/issue9/sliceutil"
)

// Dep 依赖管理
type Dep struct {
	modules []*Module
	inited  bool
	logs    *logs.Logs
}

// New 声明新的 Dep 变量
func NewDep(l *logs.Logs) *Dep {
	return &Dep{
		modules: make([]*Module, 0, 10),
		logs:    l,
	}
}

// Add 添加新模块
func (dep *Dep) Add(m ...*Module) error {
	for _, mod := range m {
		if err := dep.add(mod); err != nil {
			return err
		}
	}
	return nil
}

func (dep *Dep) add(m *Module) error {
	if sliceutil.Count(dep.modules, func(i int) bool { return dep.modules[i].name == m.name }) > 0 {
		return fmt.Errorf("模块 %s 已经存在", m.name)
	}
	dep.modules = append(dep.modules, m)

	if dep.inited {
		return m.Init("", dep.logs)
	}
	return nil
}

// Inited 是否已经初始化
func (dep *Dep) Inited() bool {
	return dep.inited
}

// Modules 模块列表
func (dep *Dep) Modules() []*Module {
	return dep.modules
}

// Tags 返回指定名称的模块的标签列表
//
// mod 表示需要查询的模块名称，如果为空，表示返回所有模块的标签列表。
//
// 返回值中键名为模块名称，键值为该模块下的标签列表。
func (dep *Dep) Tags(mod ...string) map[string][]string {
	ret := make(map[string][]string, len(mod))

	enable := func(id string) bool {
		return len(mod) == 0 ||
			sliceutil.Count(mod, func(i int) bool { return mod[i] == id }) > 0
	}

	for _, m := range dep.modules {
		if !enable(m.ID()) {
			continue
		}

		names := make([]string, 0, len(m.tags))
		for name := range m.tags {
			names = append(names, name)
		}
		sort.Strings(names)
		ret[m.ID()] = names
	}

	return ret
}

// Init 对所有的模块进行初始化操作
//
// 会进行依赖检测。若模块初始化出错，则会中断并返回出错信息。
func (dep *Dep) Init(tag string) error {
	if dep.Inited() {
		panic("已经初始化")
	}

	foundTag := tag == ""

	for _, m := range dep.modules { // 检测依赖
		if err := dep.checkDeps(m); err != nil {
			return err
		}
		if !foundTag {
			_, foundTag = m.tags[tag]
		}
	}

	if !foundTag {
		return fmt.Errorf("指定的标签 %s 不存在", tag)
	}

	for _, m := range dep.modules { // 进行初如化
		if err := dep.initModule(m, tag); err != nil {
			return err
		}
	}

	dep.inited = true
	return nil
}

// 初始化指定模块，会先初始化其依赖模块。
//
// 若该模块已经初始化，则不会作任何操作，包括依赖模块的初始化，也不会执行。
// 若 tag 不为空，表示只调用该标签下的初始化函数。
func (dep *Dep) initModule(m *Module, tag string) error {
	if tag != "" {
		if t := m.tags[tag]; t != nil {
			return t.init(dep.logs, 0)
		}
	}

	if m.Inited() {
		return nil
	}

	for _, depID := range m.deps { // 先初始化依赖项
		depMod := dep.findModule(depID)
		if depMod == nil {
			return fmt.Errorf("依赖项 %s 未找到", depID)
		}

		if err := dep.initModule(depMod, tag); err != nil {
			return err
		}
	}

	dep.logs.Info("开始初始化模块：", m.name)

	return m.Init(tag, dep.logs)
}

// 检测模块的依赖关系。比如：
// 依赖项是否存在；是否存在自我依赖等。
func (dep *Dep) checkDeps(m *Module) error {
	// 检测依赖项是否都存在
	for _, depID := range m.deps {
		if dep.findModule(depID) == nil {
			return fmt.Errorf("未找到 %s 的依赖模块 %s", m.name, depID)
		}
	}

	if dep.isDep(m.name, m.name) {
		return fmt.Errorf("存在循环依赖项:%s", m.name)
	}

	return nil
}

// m1 是否依赖 m2
func (dep *Dep) isDep(m1, m2 string) bool {
	module1 := dep.findModule(m1)
	if module1 == nil {
		return false
	}

	for _, depID := range module1.deps {
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
		if m.name == id {
			return m
		}
	}
	return nil
}
