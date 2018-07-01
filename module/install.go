// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"github.com/issue9/logs"
	"github.com/issue9/web/internal/dependency"
)

// TaskFunc 安装脚本的函数签名
type TaskFunc func() error

// Task 添加一条安装脚本
func (m *Module) Task(version, title string, fn TaskFunc) {
	if m.installs == nil {
		m.installs = make(map[string][]*task, 100)
	}

	mm, found := m.installs[version]
	if !found {
		mm = make([]*task, 0, 10)
	}

	mm = append(mm, &task{title: title, task: fn})
	m.installs[version] = mm
}

type task struct {
	title string
	task  TaskFunc
}

// GetInstall 运行当前模块的安装事件。此方法会被作为 dependency.InitFunc 被调用。
func (m *Module) GetInstall(version string) dependency.InitFunc {
	return func() error {
		logs.Infof("安装模块: %s\n", m.Name)

		hasError := false
		for _, e := range m.installs[version] {
			hasError = m.runTask(e, hasError)
		}

		if hasError {
			logs.Errorf("安装失败!\n\n")
		} else {
			logs.Infof("安装失败!\n\n")
		}
		return nil
	}
}

// 运行一条注册的事件。
//
// 若返回 true，表示继续当前模块的下一条操作，否则中止当前模块的操作。
func (m *Module) runTask(e *task, hasError bool) (err bool) {
	logs.Infof("\t%s ......", e.title)

	if hasError {
		logs.Info("[BREAK:因前一任务失败而中止]\n")
		return true
	}

	ret := e.task()

	if ret != nil {
		logs.Errorf("[FALID:%s]\n", ret.Error())
		return true
	}

	logs.Info("[OK]")
	return false
}
