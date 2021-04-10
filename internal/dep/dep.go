// SPDX-License-Identifier: MIT

// Package dep 依赖管理
package dep

import (
	"fmt"
	"log"
	"sort"

	"github.com/issue9/sliceutil"
)

// Dep 依赖管理
type Dep struct {
	ms     []*Module
	inited bool
	info   *log.Logger
}

// New 声明新的 Dep 变量
func New(info *log.Logger) *Dep {
	return &Dep{
		ms:   make([]*Module, 0, 20),
		info: info,
	}
}

// Items 返回指定名称的模块的子模块列表
//
// mod 表示需要查询的模块名称，如果为空，表示返回所有模块的子模块列表。
//
// 返回值中键名为模块名称，键值为该模块下的子模块列表。
func (d *Dep) Items(mod ...string) map[string][]string {
	ret := make(map[string][]string, len(mod))

	enable := func(id string) bool {
		return len(mod) == 0 ||
			sliceutil.Count(mod, func(i int) bool { return mod[i] == id }) > 0
	}

	for _, m := range d.ms {
		if !enable(m.ID) {
			continue
		}

		names := make([]string, 0, len(m.items))
		for name := range m.items {
			names = append(names, name)
		}
		sort.Strings(names)
		ret[m.ID] = names
	}

	return ret
}

// AddModule 添加新模块
//
// 如果所有的模块都已经初始化，则会尝试初始化 m。
func (d *Dep) AddModule(m ...*Module) error {
	for _, mod := range m {
		if err := d.addModule(mod); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dep) addModule(m *Module) error {
	if sliceutil.Count(d.ms, func(i int) bool { return d.ms[i].ID == m.ID }) > 0 {
		return fmt.Errorf("模块 %s 已经存在", m.ID)
	}
	d.ms = append(d.ms, m)

	if d.inited {
		return m.Init(d.info)
	}
	return nil
}

// Inited 是否已经初始化
func (d *Dep) Inited() bool {
	return d.inited
}

// Init 对所有的模块进行初始化操作
//
// 会进行依赖检测。若模块初始化出错，则会中断并返回出错信息。
func (d *Dep) Init(tag string) error {
	if d.Inited() {
		panic("已经初始化")
	}

	foundTag := tag == ""

	for _, m := range d.ms { // 检测依赖
		if err := d.checkDeps(m); err != nil {
			return err
		}
		if !foundTag {
			_, foundTag = m.items[tag]
		}
	}

	if !foundTag {
		return fmt.Errorf("指定的标签 %s 不存在", tag)
	}

	for _, m := range d.ms { // 进行初如化
		if err := d.initModule(m, tag); err != nil {
			return err
		}
	}

	d.inited = true
	return nil
}

// Modules 模块列表
func (d *Dep) Modules() []*Module {
	return d.ms
}

// 初始化指定模块，会先初始化其依赖模块。
//
// 若该模块已经初始化，则不会作任何操作，包括依赖模块的初始化，也不会执行。
// 若 tag 不为空，表示只调用该标签下的初始化函数。
func (d *Dep) initModule(m *Module, tag string) error {
	if tag != "" {
		m = m.items[tag]
		if m == nil {
			return nil
		}
	}

	if m.Inited() {
		return nil
	}

	for _, depID := range m.Deps { // 先初始化依赖项
		depMod := d.findModule(depID)
		if depMod == nil {
			return fmt.Errorf("依赖项 %s 未找到", depID)
		}

		if err := d.initModule(depMod, tag); err != nil {
			return err
		}
	}

	d.info.Println("开始初始化模块：", m.ID)

	return m.Init(d.info)
}

// 检测模块的依赖关系。比如：
// 依赖项是否存在；是否存在自我依赖等。
func (d *Dep) checkDeps(m *Module) error {
	// 检测依赖项是否都存在
	for _, depID := range m.Deps {
		if d.findModule(depID) == nil {
			return fmt.Errorf("未找到 %s 的依赖模块 %s", m.ID, depID)
		}
	}

	if d.isDep(m.ID, m.ID) {
		return fmt.Errorf("存在循环依赖项:%s", m.ID)
	}

	return nil
}

// m1 是否依赖 m2
func (d *Dep) isDep(m1, m2 string) bool {
	module1 := d.findModule(m1)
	if module1 == nil {
		return false
	}

	for _, depID := range module1.Deps {
		if depID == m2 {
			return true
		}

		if d.findModule(depID) != nil {
			if d.isDep(depID, m2) {
				return true
			}
		}
	}

	return false
}

func (d *Dep) findModule(id string) *Module {
	for _, m := range d.ms {
		if m.ID == id {
			return m
		}
	}
	return nil
}
