// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"log"
	"net/http"
	"sort"

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

// Module 表示模块信息
type Module struct {
	Tag
	Name        string
	Description string
	Deps        []string
	tags        map[string]*Tag
	app         *App
}

// Tag 表示与特写标签相关联的初始化函数列表。
// 依附地模块，共享模块的依赖关系。
//
// 一般是各个模块下的安装脚本使用。
type Tag struct {
	inits []*initialization
}

// 表示初始化功能的相关数据
type initialization struct {
	title string
	f     func() error
}

func (app *App) buildCoreModule(conf *webconfig.WebConfig) {
	app.coreModule = newModule(app, CoreModuleName, coreModuleDescription)
	app.modules = append(app.modules, app.coreModule)

	// 初始化静态文件处理
	for url, dir := range conf.Static {
		h := http.StripPrefix(url, http.FileServer(http.Dir(dir)))
		app.coreModule.Get(url+"{path}", h)
	}

	app.coreModule.AddService(app.scheduledService, "定时服务监控")
}

// 声明一个新的模块
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func newModule(app *App, name, desc string, deps ...string) *Module {
	return &Module{
		Name:        name,
		Description: desc,
		Deps:        deps,
		app:         app,
	}
}

// NewTag 为当前模块生成特定名称的子模块。若已经存在，则直接返回该子模块。
//
// Tag 是依赖关系与当前模块相同，但是功能完全独立的模块，
// 一般用于功能更新等操作。
func (m *Module) NewTag(tag string) *Tag {
	if m.tags == nil {
		m.tags = make(map[string]*Tag, 5)
	}

	if _, found := m.tags[tag]; !found {
		m.tags[tag] = &Tag{
			inits: make([]*initialization, 0, 5),
		}
	}

	return m.tags[tag]
}

// NewModule 声明一个新的模块
func (app *App) NewModule(name, desc string, deps ...string) *Module {
	m := newModule(app, name, desc, deps...)
	app.appendModules(m)
	return m
}

// Plugin 设置插件信息
//
// 在将模块设置为插件模式时，可以在插件的初始化函数中，采用此方法设置插件的基本信息。
func (m *Module) Plugin(name, description string, deps ...string) {
	if m.Name != "" || m.Description != "" || len(m.Deps) > 0 {
		panic("不能多次调用该函数")
	}

	m.Name = name
	m.Description = description
	m.Deps = deps
}

// AddInit 添加一个初始化函数
//
// title 该初始化函数的名称。没有则会自动生成一个序号，多个，则取第一个元素。
func (t *Tag) AddInit(f func() error, title string) *Tag {
	if t.inits == nil {
		t.inits = make([]*initialization, 0, 5)
	}

	t.inits = append(t.inits, &initialization{f: f, title: title})
	return t
}

// Modules 获取所有的模块信息
func (app *App) Modules() []*Module {
	return app.modules
}

// 收录模块到当前实例，并更 coreModule 的依赖项。
func (app *App) appendModules(modules ...*Module) {
	for _, m := range modules {
		app.modules = append(app.modules, m)

		// 让 coreModule 依赖所有模块，保证其在最后进行初始化
		app.coreModule.Deps = append(app.coreModule.Deps, m.Name)
	}
}

// Init 初始化所有的模块或是模块下指定标签名称的函数。
//
// 若指定了 tag 参数，则只初始化该名称的子模块内容。
//
// 指定 log 参数，可以输出详细的初始化步骤。
func (app *App) Init(tag string, log *log.Logger) error {
	l := func(v ...interface{}) {
		if log != nil {
			log.Println(v...)
		}
	}

	l("开始初始化模块...")

	if err := newDepencency(app, l).init(tag); err != nil {
		app.reset()
		return err
	}

	if tag == "" { // 只有模块的初始化才带路由
		all := app.Mux().All(true, true)
		if len(all) > 0 {
			l("模块加载了以下路由项：")
			for path, methods := range all {
				l(path, methods)
			}
		}
	}

	l("模块初始化完成！")

	return nil
}

// 重置状态为未初始化时的状态
func (app *App) reset() {
	app.router.Clean()

	app.Stop() // 先停止停止服务
	app.services = app.services[:0]
}

// Stop 停止服务
func (app *App) Stop() {
	for _, srv := range app.services {
		srv.Stop()
	}
}

// Tags 返回所有的子模块名称
func (app *App) Tags() []string {
	tags := make([]string, 0, len(app.modules)*2)

	for _, m := range app.modules {
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
