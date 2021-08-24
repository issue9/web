// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"log"
	"sort"

	"github.com/issue9/sliceutil"
)

// InitModules 触发所有模块下指定名称的函数
func (srv *Server) InitModules(tag string) error {
	if tag == "" {
		panic("参数  tag 不能为空")
	}

	for _, m := range srv.modules { // 检测依赖
		if err := srv.checkDeps(m); err != nil {
			return err
		}
	}

	l := srv.Logs().INFO()

	// 日志不需要标出文件位置。
	flags := l.Flags()
	l.SetFlags(log.Ldate | log.Lmicroseconds)

	l.Printf("开始初始化模块中的 %s...\n", tag)
	for _, m := range srv.modules { // 进行初如化
		if err := srv.initModule(m, l, tag); err != nil {
			return err
		}
	}
	l.Print("初始化完成！\n\n")

	l.SetFlags(flags)

	return nil
}

func (srv *Server) initModule(m *Module, l *log.Logger, tag string) error {
	for _, depID := range m.deps { // 先初始化依赖项
		depMod := srv.findModule(depID)
		if depMod == nil {
			return fmt.Errorf("模块 %s 依赖项 %s 未找到", m.id, depID)
		}

		if err := srv.initModule(depMod, l, tag); err != nil {
			return err
		}
	}

	l.Println(m.id, "...")

	err := m.Tag(tag).init(l)
	if err != nil {
		l.Printf("%s [FAIL:%s]\n\n", m.id, err.Error())
	} else {
		l.Printf("%s [OK]\n\n", m.id)
	}

	return err
}

// 检测模块的依赖关系。比如：
// 依赖项是否存在；是否存在自我依赖等。
func (srv *Server) checkDeps(m *Module) error {
	// 检测依赖项是否都存在
	for _, depID := range m.deps {
		if srv.findModule(depID) == nil {
			return fmt.Errorf("未找到 %s 的依赖模块 %s", m.id, depID)
		}
	}

	if srv.isDep(m.id, m.id) {
		return fmt.Errorf("%s 循环依赖自身", m.id)
	}

	return nil
}

// m1 是否依赖 m2
func (srv *Server) isDep(m1, m2 string) bool {
	mod1 := srv.findModule(m1)
	if mod1 == nil {
		return false
	}

	for _, depID := range mod1.deps {
		if depID == m2 {
			return true
		}

		if srv.findModule(depID) != nil {
			if srv.isDep(depID, m2) {
				return true
			}
		}
	}

	return false
}

func (srv *Server) findModule(id string) *Module {
	for _, m := range srv.modules {
		if m.id == id {
			return m
		}
	}
	return nil
}

func (srv *Server) Tags() []string {
	tags := make([]string, 0, 100)
	for _, m := range srv.modules {
		tags = append(tags, m.Tags()...)
	}
	size := sliceutil.Unique(tags, func(i, j int) bool { return tags[i] == tags[j] })
	tags = tags[:size]
	sort.Strings(tags)
	return tags
}
