// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package modules 处理模块信息
package modules

import (
	"net/http"
	"path"

	"github.com/issue9/mux"

	"github.com/issue9/web/internal/app/webconfig"
	"github.com/issue9/web/module"
)

// CoreModuleName 框架名
const CoreModuleName = "web-core"

// Modules 模块管理
type Modules struct {
	modules []*module.Module
	router  *mux.Prefix
}

// New 声明 Modules 变量
func New(cap int, mux *mux.Mux, conf *webconfig.WebConfig) (*Modules, error) {
	ms := &Modules{
		router:  mux.Prefix(conf.Root),
		modules: make([]*module.Module, 0, cap),
	}

	// 默认的模块
	m := ms.NewModule(CoreModuleName, CoreModuleName)

	// 初始化静态文件处理
	for url, dir := range conf.Static {
		pattern := path.Join(conf.Root, url+"{path}")
		fs := http.FileServer(http.Dir(dir))
		m.Get(pattern, http.StripPrefix(url, fs))
	}

	// 在初始化模块之前，先加载插件
	if conf.Plugins != "" {
		modules, err := loadPlugins(conf.Plugins, ms.router)
		if err != nil {
			return nil, err
		}
		ms.modules = append(ms.modules, modules...)
	}

	return ms, nil
}

// NewModule 声明一个新的模块
func (ms *Modules) NewModule(name, desc string, deps ...string) *module.Module {
	m := module.New(module.TypeModule, name, desc, deps...)
	ms.modules = append(ms.modules, m)
	return m
}

// Init 初如化插件
func (ms *Modules) Init(tag string) error {
	return newDepencency(ms.modules).init(tag, ms.router)
}

// Modules 获取所有的模块信息
func (ms *Modules) Modules() []*module.Module {
	return ms.modules
}
