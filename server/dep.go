// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"log"
	"sort"

	"github.com/issue9/sliceutil"
)

// InitModules 触发所有模块下指定名称的函数
//
// module 表示需要初始化的模块 ID，如果为这，表示所有已经添加的模块。
func (srv *Server) InitModules(action string, module ...string) error {
	if action == "" {
		panic("参数  action 不能为空")
	}

	for _, m := range srv.modules { // 检测依赖
		if err := srv.checkDeps(m); err != nil {
			return err
		}
	}

	needInit := func(id string) bool {
		return len(module) == 0 ||
			sliceutil.Count(module, func(i int) bool { return module[i] == id }) >= 0
	}

	l := srv.Logs().INFO()

	// 日志不需要标出文件位置。
	flags := l.Flags()
	l.SetFlags(log.Ldate | log.Lmicroseconds)

	l.Printf("开始初始化模块中的 %s...\n", action)
	for _, m := range srv.modules { // 进行初如化
		if !needInit(m.id) {
			continue
		}

		if err := srv.initModule(m, l, action); err != nil {
			return err
		}
	}
	l.Print("初始化完成！\n\n")

	l.SetFlags(flags)

	return nil
}

func (srv *Server) initModule(m *Module, l *log.Logger, action string) error {
	for _, depID := range m.deps { // 先初始化依赖项
		depMod := srv.findModule(depID)
		if depMod == nil {
			return fmt.Errorf("模块 %s 依赖项 %s 未找到", m.id, depID)
		}

		if err := srv.initModule(depMod, l, action); err != nil {
			return err
		}
	}

	l.Println(m.id, "...")

	err := m.Action(action).init(l)
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

func (srv *Server) Actions() []string {
	actions := make([]string, 0, 100)
	for _, m := range srv.modules {
		actions = append(actions, m.Actions()...)
	}
	size := sliceutil.Unique(actions, func(i, j int) bool { return actions[i] == actions[j] })
	actions = actions[:size]
	sort.Strings(actions)
	return actions
}
