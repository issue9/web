// SPDX-License-Identifier: MIT

// Package dep 依赖管理
package dep

import (
	"errors"
	"fmt"
	"log"
)

var (
	// ErrInited 当模块被多次初始化时返回此错误
	ErrInited = errors.New("模块已经初始化")

	// ErrModExists 已经添加了相同 ID 的模块实现
	ErrModExists = errors.New("模块已经存在")
)

// Dep 依赖管理
type Dep struct {
	ms     []Module
	inited bool
	info   *log.Logger
}

// New 声明新的 Dep 变量
func New(info *log.Logger) *Dep {
	return &Dep{
		ms:   make([]Module, 0, 20),
		info: info,
	}
}

// AddModule 添加新模块
func (d *Dep) AddModule(m Module) error {
	for _, mod := range d.ms {
		if mod.ID() == m.ID() {
			return ErrModExists
		}
	}
	d.ms = append(d.ms, m)

	if d.inited {
		return m.Init(d.info)
	}

	return nil
}

// Init 对所有的模块进行初始化操作
//
// 会进行依赖检测。若模块初始化出错，则会中断并返回出错信息。
func (d *Dep) Init() error {
	if d.inited {
		return ErrInited
	}

	for _, m := range d.ms { // 检测依赖
		if err := d.checkDeps(m); err != nil {
			return err
		}
	}

	for _, m := range d.ms { // 进行初如化
		if err := d.initModule(m, d.info); err != nil {
			return err
		}
	}

	d.inited = true
	return nil
}

// Modules 模块列表
func (d *Dep) Modules() []Module {
	return d.ms
}

// 初始化指定模块，会先初始化其依赖模块。
//
// 若该模块已经初始化，则不会作任何操作，包括依赖模块的初始化，也不会执行。
// 若 tag 不为空，表示只调用该标签下的初始化函数。
func (d *Dep) initModule(m Module, info *log.Logger) error {
	if m.Inited() {
		return nil
	}

	for _, depID := range m.Deps() { // 先初始化依赖项
		depMod := d.findModule(depID)
		if depMod == nil {
			return fmt.Errorf("依赖项 %s 未找到", depID)
		}

		if err := d.initModule(depMod, info); err != nil {
			return err
		}
	}

	info.Println("开始初始化模块：", m.ID())

	return m.Init(info)
}

// 检测模块的依赖关系。比如：
// 依赖项是否存在；是否存在自我依赖等。
func (d *Dep) checkDeps(m Module) error {
	// 检测依赖项是否都存在
	for _, depID := range m.Deps() {
		if d.findModule(depID) == nil {
			return fmt.Errorf("未找到 %s 的依赖模块 %s", m.ID(), depID)
		}
	}

	if d.isDep(m.ID(), m.ID()) {
		return fmt.Errorf("存在循环依赖项:%s", m.ID())
	}

	return nil
}

// m1 是否依赖 m2
func (d *Dep) isDep(m1, m2 string) bool {
	module1 := d.findModule(m1)
	if module1 == nil {
		return false
	}

	for _, depID := range module1.Deps() {
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

func (d *Dep) findModule(id string) Module {
	for _, m := range d.ms {
		if m.ID() == id {
			return m
		}
	}
	return nil
}
