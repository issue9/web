// SPDX-License-Identifier: MIT

package module

import (
	"errors"
	"fmt"
	"sort"

	"github.com/issue9/logs/v2"
	"github.com/issue9/sliceutil"
)

// ErrInited 当模块被多次初始化时返回此错误
var ErrInited = errors.New("模块已经初始化")

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
func (dep *Dep) Inited() bool { return dep.inited }

// Modules 模块列表
func (dep *Dep) Modules() []*Module { return dep.modules }

// Tags 返回所有模块的标签列表
//
// 如果要查看单个模块的，可调用 Module.Tags() 方法。
func (dep *Dep) Tags() []string {
	tags := make([]string, 0, 100)
	for _, m := range dep.modules {
		tags = append(tags, m.Tags()...)
	}
	index := sliceutil.Unique(tags, func(i, j int) bool { return tags[i] == tags[j] })
	tags = tags[:index]
	sort.Strings(tags)
	return tags
}

// Init 对所有的模块进行初始化操作
//
// 会进行依赖检测。若模块初始化出错，则会中断并返回出错信息。
func (dep *Dep) Init(tag string) error {
	if dep.Inited() {
		return ErrInited
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

	dep.inited = tag == "" // tag 不为空时，不设置 inited
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
		return fmt.Errorf("%s 循环依赖自身", m.name)
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
