// SPDX-License-Identifier: MIT

// Package module 提供模块管理的相关功能
package module

import (
	"log"
	"sort"
	"time"

	"github.com/issue9/web/config"
)

// Module 表示模块信息
type Module struct {
	Tag
	Name        string
	Description string
	Deps        []string
	tags        map[string]*Tag
	ms          *Modules
}

// Tag 表示与特写标签相关联的初始化函数列表。
// 依附于模块，共享模块的依赖关系。
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

func (srv *Modules) newModule(name, desc string, deps ...string) *Module {
	return &Module{
		Name:        name,
		Description: desc,
		Deps:        deps,
		ms:          srv,
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
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func (srv *Modules) NewModule(name, desc string, deps ...string) *Module {
	m := srv.newModule(name, desc, deps...)
	srv.modules = append(srv.modules, m)
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

// InitModules 初始化模块下指定标签名称的函数
//
// 若指定了 tag 参数，则只初始化该名称的子模块内容。
func (srv *Modules) InitModules(tag string, info *log.Logger) error {
	flag := info.Flags()
	info.SetFlags(0)
	defer info.SetFlags(flag)

	info.Println("开始初始化模块...")

	if err := newDepencency(srv.modules, info).init(tag); err != nil {
		return err
	}

	if tag == "" { // 只有模块的初始化才带路由
		all := srv.ctxServer.Router().Mux().All(true, true)
		if len(all) > 0 {
			info.Println("模块加载了以下路由项：")
			for path, methods := range all {
				info.Println(path, methods)
			}
		}
	}

	info.Println("模块初始化完成！")

	return nil
}

// Tags 返回所有的子模块名称
//
// 键名为模块名称，键值为该模块下的标签列表
func (srv *Modules) Tags() map[string][]string {
	ret := make(map[string][]string, len(srv.modules)*2)

	for _, m := range srv.modules {
		tags := make([]string, 0, len(m.tags))
		for k := range m.tags {
			tags = append(tags, k)
		}
		sort.Strings(tags)
		ret[m.Name] = tags
	}

	return ret
}

// Modules 当前系统使用的所有模块信息
func (srv *Modules) Modules() []*Module {
	return srv.modules
}

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddCron(title string, f JobFunc, spec string, delay bool) {
	m.AddInit(func() error {
		return m.ms.Scheduled().Cron(title, f, spec, delay)
	}, "注册计划任务"+title)
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// imm 是否立即执行一次该任务；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddTicker(title string, f JobFunc, dur time.Duration, imm, delay bool) {
	m.AddInit(func() error {
		return m.ms.Scheduled().Tick(title, f, dur, imm, delay)
	}, "注册计划任务"+title)
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddAt(title string, f JobFunc, spec string, delay bool) {
	m.AddInit(func() error {
		return m.ms.Scheduled().At(title, f, spec, delay)
	}, "注册计划任务"+title)
}

// RegisterConfig 注册配置项
func (m *Module) RegisterConfig(id, config string, v interface{}, f config.UnmarshalFunc, notify func()) error {
	return m.ms.config.Register(id, config, v, f, notify)
}

// RefreshConfig 刷新指定 ID 的配置项
func (m *Module) RefreshConfig(id string) error {
	return m.ms.config.Refresh(id)
}
