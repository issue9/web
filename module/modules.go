// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"log"
	"net/http"
	"sort"

	"github.com/issue9/middleware"
	"github.com/issue9/mux/v2"

	"github.com/issue9/web/internal/webconfig"
)

const (
	// CoreModuleName 框架自带的模块名称
	//
	// 该模块会在所有模块初始化之后，进行最后的初始化操作，包括了以下内容：
	// - 配置文件中指定的静态文件服务内容 static；
	// - 所有模块注册的服务，也由此模块负责启动。
	CoreModuleName        = "web-core"
	coreModuleDescription = "web 框架自带的模块，包括启动服务等，最后加载运行"
)

// Modules 模块管理
//
// 负责模块的初始化工作，包括路由的加载等。
type Modules struct {
	middleware.Manager
	modules  []*Module
	router   *mux.Prefix
	services []*Service

	// 框架自身用到的模块，除了配置文件中的静态文件服务之外，
	// 还有所有服务的注册启动等，其初始化包含了启动服务等。
	coreModule *Module
}

// NewModules 声明 Modules 变量
func NewModules(conf *webconfig.WebConfig) (*Modules, error) {
	mux := mux.New(conf.DisableOptions, conf.DisableHead, false, nil, nil)
	ms := &Modules{
		Manager:  *middleware.NewManager(mux),
		modules:  make([]*Module, 0, 100),
		router:   mux.Prefix(conf.Root),
		services: make([]*Service, 0, 100),
	}

	ms.buildCoreModule(conf)

	if conf.Plugins != "" {
		if err := ms.loadPlugins(conf.Plugins); err != nil {
			return nil, err
		}
	}

	return ms, nil
}

// 收录模块到当前实例，并更 coreModule 的依赖项。
func (ms *Modules) appendModules(modules ...*Module) {
	for _, m := range modules {
		ms.modules = append(ms.modules, m)

		// 让 coreModule 依赖所有模块，保证其在最后进行初始化
		ms.coreModule.Deps = append(ms.coreModule.Deps, m.Name)
	}
}

func (ms *Modules) buildCoreModule(conf *webconfig.WebConfig) {
	ms.coreModule = newModule(ms, CoreModuleName, coreModuleDescription)
	ms.modules = append(ms.modules, ms.coreModule)

	// 初始化静态文件处理
	for url, dir := range conf.Static {
		h := http.StripPrefix(url, http.FileServer(http.Dir(dir)))
		ms.coreModule.Get(url+"{path}", h)
	}
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
	if err := newDepencency(ms, log).init(tag); err != nil {
		ms.reset()
		return err
	}

	return nil
}

// 重置状态为未初始化时的状态
func (ms *Modules) reset() {
	ms.router.Clean()

	ms.Stop() // 先停止停止服务
	ms.services = ms.services[:0]

	for _, m := range ms.modules {
		m.inited = false
	}
}

// Modules 获取所有的模块信息
func (ms *Modules) Modules() []*Module {
	return ms.modules
}

// Services 返回所有的服务列表
func (ms *Modules) Services() []*Service {
	return ms.services
}

// Stop 停止服务
func (ms *Modules) Stop() {
	for _, srv := range ms.services {
		srv.Stop()
	}
}

// Tags 返回所有的子模块名称
func (ms *Modules) Tags() []string {
	tags := make([]string, 0, len(ms.modules)*2)

	for _, m := range ms.modules {
	LOOP:
		for k := range m.tags {
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
