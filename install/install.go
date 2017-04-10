// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package install 一个简短的安装日志显示功能。
// 可以在控制台细粒度地显示安装进程。
//
// 正常安装内容，则显示如下内容：
//  安装模块users...
//      install  users [ok]
//      install  email [ok]
//  完成安装
//
// 安装出错，则显示如下内容：
//  安装模块users...
//      install  users [ok]
//      install  email [falid:message]
//  完成失败
//
// 使用方法：
//  声明一个安装模块
//  i := install.New("admin")
//  defer i.Done()
//
//  i.Event("创建表:users", func()*install.Result{
//      return nil
//  })
//
//  i.Event("创建表:users2", func()*install.Result{
//      err := db.Insert(...)
//      return install.ReturnError(err)
//  })
//
//
//  // 执行安装
//  install.Install()
package install

import (
	"github.com/issue9/term/colors"
	"github.com/issue9/web/modules"
)

var defaultModules = modules.New()

// Module 声明了一个用于安装的模块。
// 所有的安装事件都可以向模块注册，模块会在适当的时候进行初始化。
type Module struct {
	name     string
	deps     []string
	hasError bool
	events   []*event
}

type event struct {
	title string
	fn    func() *Return
}

// New 输出模块开始安装的信息。
func New(module string, deps ...string) *Module {
	return &Module{
		name:     module,
		deps:     deps,
		hasError: false,
		events:   make([]*event, 0, 10),
	}
}

// Event 为当前模块注册安装事件。
//
// name 事件名称。
// fn 事件的处理函数。
func (i *Module) Event(title string, fn func() *Return) {
	i.events = append(i.events, &event{
		title: title,
		fn:    fn,
	})
}

// Events 批量注册事件。
//
// events 为一个 map，键名为事件名称，键值为事件处理函数。
//
// NOTE:事件名称若相同，会相互覆盖。
func (i *Module) Events(events map[string]func() *Return) {
	for title, fn := range events {
		i.Event(title, fn)
	}
}

// Done 完成当前安装模块的所有事件注册
func (i *Module) Done() error {
	return defaultModules.New(i.name, i.run, i.deps...)
}

// 运行当前模块的安装事件。
//
// NOTE: 此函数会作为 modules 包中的模块初始化函数，
// 在初始化所有模块时，根据依赖顺序被调用。
func (i *Module) run() error {
	colorPrint(colors.Green, "安装模块:")
	colorPrintf(colors.Default, "[%v]\n", i.name)

	for _, e := range i.events {
		i.runEvent(e)
	}

	if i.hasError {
		colorPrint(colors.Red, "安装失败!\n\n")
	} else {
		colorPrint(colors.Green, "安装完成!\n\n")
	}
	return nil
}

// 运行一条注册的事件。
func (i *Module) runEvent(e *event) {
	colorPrint(colors.Default, "\t", e.title, "......")

	ret := e.fn()

	if ret != nil && ret.typ == typeFailed {
		i.hasError = true
		colorPrintf(colors.Red, "[FALID:%v]\n", ret.message)
		return
	}

	colorPrint(colors.Green, "[OK")
	if ret != nil && len(ret.message) > 0 {
		colorPrint(colors.Magenta, ":")
		colorPrint(colors.Magenta, ret.message)
	}
	colorPrintln(colors.Green, "]")
}

// Install 安装各个模块
func Install() error {
	return defaultModules.Init()
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
