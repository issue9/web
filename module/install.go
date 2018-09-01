// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"net/http"

	"github.com/issue9/term/colors"
	"github.com/issue9/web/internal/dependency"
)

const (
	colorDefault = colors.Default
	colorInfo    = colors.Magenta
	colorError   = colors.Red
	colorSuccess = colors.Green
)

// TaskFunc 安装脚本的函数签名
type TaskFunc func() error

type task struct {
	title string
	task  TaskFunc
}

type message string

// NewMessage 声明一条 message 类型的错误信息
//
// 返回内容并不是一个真正的错误，则是在某些时候需要在安装完成之后，
// 反馈一些文字信息，则需要用此函数进行包装。
func NewMessage(msg string) error {
	return message(msg)
}

func (msg message) Error() string {
	return string(msg)
}

// NewVersion 为当前模块添加某一版本号下的安装脚本。
func (m *Module) NewVersion(version string) *Module {
	return m.NewTag(version)
}

// NewTag 为当前模块添加某一版本号下的安装脚本。
func (m *Module) NewTag(tag string) *Module {
	if _, found := m.Tags[tag]; !found {
		m.Tags[tag] = &Module{
			Type:        TypeModule,
			Name:        "",
			Deps:        nil,
			Description: "",
			Routes:      make(map[string]map[string]http.Handler, 10),
			Inits:       make([]func() error, 0, 5),
			Installs:    make([]*task, 0, 10),
		}
	}

	return m.Tags[tag]
}

// Task 添加一条安装脚本
func (m *Module) Task(title string, fn TaskFunc) {
	m.Installs = append(m.Installs, &task{title: title, task: fn})
}

// GetInstall 运行当前模块的安装事件。此方法会被作为 dependency.InitFunc 被调用。
func (m *Module) GetInstall(tag string) dependency.InitFunc {
	return func() error {
		colorPrintf(colorDefault, "安装模块: %s\n", m.Name)

		t, found := m.Tags[tag]
		if !found {
			colorPrint(colorInfo, "不存在此版本的安装脚本!\n\n")
			return nil
		}

		hasError := false
		for _, e := range t.Installs {
			hasError = m.runTask(e, hasError)
		}

		if hasError {
			colorPrint(colorError, "安装失败!\n\n")
		} else {
			colorPrint(colorSuccess, "安装完成!\n\n")
		}
		return nil
	}
}

// 运行一条安装的事件。
//
// 若返回 true，表示继续当前模块的下一条操作，否则中止当前模块的操作。
//
// 返回值表示当前执行是否出错，若出错返回 true
func (m *Module) runTask(e *task, hasError bool) bool {
	colorPrintf(colorInfo, "\t%s ......", e.title)

	if hasError {
		colorPrintln(colorError, "[BREAK:因前一任务失败而中止]")
		return true
	}

	err := e.task()

	if err == nil {
		colorPrintln(colorSuccess, "[OK]")
		return false
	}

	if msg, ok := err.(message); ok {
		colorPrintf(colorSuccess, "[OK:%s]\n", msg)
		return false
	}

	colorPrintf(colorError, "[FALID:%s]\n", err.Error())
	return true
}

// 打印指定颜色的字符串
func colorPrint(color colors.Color, msg ...interface{}) {
	colors.Print(color, colors.Default, msg...)
}

// 打印指定颜色的字符串并换行
func colorPrintln(color colors.Color, msg ...interface{}) {
	colors.Println(color, colors.Default, msg...)
}

// 打印指定颜色的字符串
func colorPrintf(color colors.Color, msg string, vals ...interface{}) {
	colors.Printf(color, colors.Default, msg, vals...)
}
