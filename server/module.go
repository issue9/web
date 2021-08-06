// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"plugin"

	"github.com/issue9/web/dep"
)

// ModuleFuncName 插件中的用于获取模块信息的函数名
//
// NOTE: 必须为可导出的函数名称
const ModuleFuncName = "InitModule"

// ModuleFunc 安装插件的函数签名
type ModuleFunc func(*Server) error

type Module struct {
	*dep.Module
	fs.FS
}

// NewModule 声明一个新的模块
//
// id 模块名称，需要全局唯一；
// version 模块的版本信息；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func (srv *Server) NewModule(id, version, desc string, deps ...string) (*Module, error) {
	m, err := srv.dep.NewModule(id, version, desc, deps...)
	if err != nil {
		return nil, err
	}

	sub, err := fs.Sub(srv.fs, id)
	if err != nil {
		return nil, err
	}

	return &Module{
		Module: m,
		FS:     sub,
	}, nil
}

// Tags 返回所有的子模块名称
//
// 键名为模块名称，键值为该模块下的标签列表。
func (srv *Server) Tags() []string { return srv.dep.Tags() }

// Modules 当前系统使用的所有模块信息
func (srv *Server) Modules() []*dep.Module { return srv.dep.Modules() }

// InitTag 初始化模块下的子标签
func (srv *Server) InitModules(tag string) error {
	return srv.dep.Init(srv.Logs().INFO(), tag)
}

// LoadPlugins 加载所有的插件
//
// 如果 glob 为空，则不会加载任何内容，返回空值
func (srv *Server) LoadPlugins(glob string) error {
	fs, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	for _, path := range fs {
		if err := srv.LoadPlugin(path); err != nil {
			return err
		}
	}

	return nil
}

// LoadPlugin 将指定的插件当作模块进行加载
//
// path 为插件的路径；
//
// 插件必须是以 buildmode=plugin 的方式编译的，且要求其引用的 github.com/issue9/web
// 版本与当前的相同。
// LoadPlugin 会在插件中查找固定名称和类型的函数名（参考 ModuleFunc 和 ModuleFuncName），
// 如果存在，会调用该方法将插件加载到 Server 对象中，否则返回相应的错误信息。
func (srv *Server) LoadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return err
	}

	symbol, err := p.Lookup(ModuleFuncName)
	if err != nil {
		return err
	}

	if install, ok := symbol.(func(*Server) error); ok {
		return install(srv)
	}
	return fmt.Errorf("插件 %s 未找到安装函数", path)
}
