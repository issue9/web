// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package install 一个简短的安装日志显示功能。
// 可以在控制台细粒度地显示安装进程。
//
// 正常安装内容，则显示如下内容：
//  安装模块 users...
//      install  users [ok]
//      install  email [ok]
//  完成安装
//
// 安装出错，则显示如下内容：
//  安装模块 users...
//      install  users [ok]
//      install  email [falid:message]
//  安装失败
//
// 使用方法：
//  声明一个安装模块
//  i := install.Get("1.0").New("admin")
//
//  i.Task("创建表:users", func()*install.Result{
//      return nil
//  })
//
//  i.Task("创建表:users2", func()*install.Result{
//      err := db.Insert(...)
//      return install.ReturnError(err)
//  })
//
//
//  // 执行安装
//  install.Install("1.0")
package install

import (
	"github.com/issue9/term/colors"
	"github.com/issue9/web/internal/dependency"
)

const (
	colorDefault = colors.Default
	colorInfo    = colors.Magenta
	colorError   = colors.Red
	colorSuccess = colors.Green
)

var versions = map[string]*Version{}

// Version 某一版本下的安装信息
type Version struct {
	modules []*Module
}

// Module 声明了一个用于安装的模块。
// 所有的安装事件都可以向模块注册，模块会在适当的时候进行初始化。
type Module struct {
	name     string
	deps     []string
	hasError bool
	tasks    []*task
}

type task struct {
	title string
	fn    func() *Return
}

// Get 获取指定版本相关的安装信息
func Get(version string) *Version {
	if v, found := versions[version]; found {
		return v
	}

	v := &Version{
		modules: make([]*Module, 0, 10),
	}

	versions[version] = v
	return v
}

// New 输出模块开始安装的信息。
func (v *Version) New(module string, deps ...string) *Module {
	m := &Module{
		name:     module,
		deps:     deps,
		hasError: false,
		tasks:    make([]*task, 0, 10),
	}

	v.modules = append(v.modules, m)

	return m
}

// Task 为当前模块添加任务。
//
// name 事件名称。
// fn 事件的处理函数。
func (m *Module) Task(title string, fn func() *Return) *Module {
	m.tasks = append(m.tasks, &task{
		title: title,
		fn:    fn,
	})

	return m
}

// 运行当前模块的安装事件。此方法会被作为 dependency.InitFunc 被调用。
func (m *Module) run() error {
	colorPrint(colorSuccess, "安装模块:")
	colorPrintf(colorDefault, "[%v]\n", m.name)

	for _, e := range m.tasks {
		m.runEvent(e)
	}

	if m.hasError {
		colorPrint(colorError, "安装失败!\n\n")
	} else {
		colorPrint(colorSuccess, "安装完成!\n\n")
	}
	return nil
}

// 运行一条注册的事件。
//
// 若返回 true，表示继承当前模块的下一条操作，否则中止当前模块的操作。
func (m *Module) runEvent(e *task) {
	colorPrint(colorDefault, "\t", e.title, "......")

	if m.hasError {
		colorPrintf(colorInfo, "[BREAK:因前一任务失败而中止]\n")
		return
	}

	ret := e.fn()

	if ret != nil && ret.typ == typeFailed {
		m.hasError = true
		colorPrintf(colorError, "[FALID:%s]\n", ret.message)
		return
	}

	colorPrint(colorSuccess, "[OK")
	if ret != nil && len(ret.message) > 0 {
		colorPrint(colorInfo, ":")
		colorPrint(colorInfo, ret.message)
	}
	colorPrintln(colorSuccess, "]")
}

// Install 安装各个模块
func Install(version string) error {
	v, found := versions[version]
	if !found {
		return nil
	}

	dep := dependency.New()
	for _, m := range v.modules {
		dep.Add(m.name, m.run, m.deps...)
	}

	return dep.Init()
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