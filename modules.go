// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import "github.com/issue9/logs"
import "github.com/issue9/web/modules"

var defaultModules = modules.New()

// NewModule 注册一个新的模块。
//
// name 为模块名称；
// init 当前模块的初始化函数；
// deps 模块的依赖模块，这些模块在初始化时，会先于 name 初始化始。
func NewModule(name string, init modules.Init, deps ...string) {
	err := defaultModules.New(name, init, deps...)

	// 注册模块时出错，直接退出。
	if err != nil {
		logs.Fatal(err)
	}
}
