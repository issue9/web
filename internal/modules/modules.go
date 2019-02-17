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

	"github.com/issue9/middleware"
	"github.com/issue9/mux/v2"

	"github.com/issue9/web/internal/webconfig"
	"github.com/issue9/web/module"
)

const (
	// CoreModuleName 模块名称
	CoreModuleName        = "web-core"
	coreModuleDescription = "web 框架自带的模块，包括启动服务等，最后加载运行"
)

// Modules 模块管理
//
// 负责模块的初始化工作，包括路由的加载等。
type Modules struct {
	middleware.Manager

	// 框架自身用到的模块，除了配置文件中的静态文件服务之外，
	// 还有所有服务的注册启动等，其初始化包含了启动服务等。
	coreModule *module.Module

	modules  []*module.Module
	router   *mux.Prefix
	services []*module.Service
}

// New 声明 Modules 变量
func New(conf *webconfig.WebConfig) (*Modules, error) {
	mux := mux.New(conf.DisableOptions, conf.DisableHead, false, nil, nil)
	ms := &Modules{
		Manager:  *middleware.NewManager(mux),
		modules:  make([]*module.Module, 0, 100),
		router:   mux.Prefix(conf.Root),
		services: make([]*module.Service, 0, 100),
	}

	ms.buildCoreModule(conf)

	// 在初始化模块之前，先加载插件
	if conf.Plugins != "" {
		modules, err := loadPlugins(conf.Plugins)
		if err != nil {
			return nil, err
		}
		ms.appendModules(modules...)
	}

	return ms, nil
}

func (ms *Modules) appendModules(modules ...*module.Module) {
	for _, m := range modules {
		ms.modules = append(ms.modules, m)

		// 让 coreModule 依赖所有模块，保证其在最后进行初始化
		ms.coreModule.Deps = append(ms.coreModule.Deps, m.Name)
	}
}

func (ms *Modules) buildCoreModule(conf *webconfig.WebConfig) {
	ms.coreModule = module.New(module.TypeModule, CoreModuleName, coreModuleDescription)
	ms.modules = append(ms.modules, ms.coreModule)

	// 初始化静态文件处理
	for url, dir := range conf.Static {
		pattern := path.Join(conf.Root, url+"{path}")
		ms.coreModule.Get(pattern, http.StripPrefix(url, http.FileServer(http.Dir(dir))))
	}

	ms.coreModule.AddInit(func() error {
		for _, srv := range ms.services {
			srv.Run()
		}
		return nil
	}, "启动常驻服务")
}

// NewModule 声明一个新的模块
func (ms *Modules) NewModule(name, desc string, deps ...string) *module.Module {
	m := module.New(module.TypeModule, name, desc, deps...)
	ms.appendModules(m)
	return m
}

// Mux 返回相关的 mux.Mux 实例
func (ms *Modules) Mux() *mux.Mux {
	return ms.router.Mux()
}

// Init 初始化插件
//
// 若指定了 tag 参数，则只初始化该名称的子模块内容。
//
// 指定 log 参数，可以输出详细的初始化步骤。
func (ms *Modules) Init(tag string, log *log.Logger) error {
	return newDepencency(ms, log).init(tag)
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
