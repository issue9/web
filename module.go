// SPDX-License-Identifier: MIT

package web

// Modules 当前系统使用的所有模块信息
func Modules() []*Module {
	return defaultServer.Modules()
}

// Tags 按模块返回各自的标签列表
func Tags() map[string][]string {
	return defaultServer.Tags()
}

// NewModule 声明一个新的模块
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func NewModule(name, desc string, deps ...string) *Module {
	return defaultServer.NewModule(name, desc, deps...)
}

// InitModules 初始化所有的模块或是模块下指定标签名称的函数。
//
// 若指定了 tag 参数，则只初始化该名称的子模块内容。
func InitModules(tag string) error {
	return defaultServer.InitModules(tag, INFO())
}
