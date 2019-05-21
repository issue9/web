// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import "github.com/issue9/web/module"

var modules *module.Modules

// Modules 当前系统使用的所有模块信息
func Modules() []*Module {
	return modules.Modules()
}

// NewModule 声明一个新的模块
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func NewModule(name, desc string, deps ...string) *Module {
	return modules.New(name, desc, deps...)
}

// InitModules 初始化所有的模块或是模块下指定标签名称的函数。
//
// 若指定了 tag 参数，则只初始化该名称的子模块内容。
func InitModules(tag string) error {
	return modules.Init(tag, INFO())
}
