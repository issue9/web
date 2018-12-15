// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package modules 处理模块信息
package modules

import (
	"log"
	"net/http"
	"path"
	"sort"

	"github.com/issue9/mux"

	"github.com/issue9/web/internal/modules/dep"
	"github.com/issue9/web/internal/webconfig"
	"github.com/issue9/web/module"
)

const (
	coreModuleName        = "web-core"
	coreModuleDescription = "web-core"
)

// Modules 模块管理
type Modules struct {
	modules []*module.Module
	router  *mux.Prefix
}

// New 声明 Modules 变量
func New(mux *mux.Mux, conf *webconfig.WebConfig) (*Modules, error) {
	ms := &Modules{
		router:  mux.Prefix(conf.Root),
		modules: make([]*module.Module, 0, 100),
	}

	// 默认的模块
	// 数组的第一个元素，且没有依赖，可以确保是第一个被初始化的元素。
	m := ms.NewModule(coreModuleName, coreModuleDescription)

	// 初始化静态文件处理
	// BUG(caixw): http.FileServer 无法自定义 404 等错误的行为。
	// https://github.com/issue9/web/issues/4
	for url, dir := range conf.Static {
		pattern := path.Join(conf.Root, url+"{path}")
		fs := http.FileServer(http.Dir(dir))
		m.Get(pattern, http.StripPrefix(url, fs))
	}

	// 在初始化模块之前，先加载插件
	if conf.Plugins != "" {
		modules, err := loadPlugins(conf.Plugins)
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
//
// 若指定了 tag 参数，则只初始化该名称的子模块内容。
//
// 指定 log 参数，可以输出详细的初始化步骤。
func (ms *Modules) Init(tag string, log *log.Logger) error {
	return dep.Do(ms.modules, ms.router, log, tag)
}

// Modules 获取所有的模块信息
func (ms *Modules) Modules() []*module.Module {
	return ms.modules
}

// Tags 返回所有的子模块名称
func (ms *Modules) Tags() []string {
	tags := make([]string, 0, len(ms.modules)*2)

	for _, m := range ms.modules {
	LOOP:
		for k := range m.Tags {
			for _, v := range tags {
				if v == k {
					continue LOOP
				}
			}
			tags = append(tags, k)
		}
	}

	sort.Strings(tags)

	return tags
}