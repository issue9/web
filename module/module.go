// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package module 模块的管理
package module

import (
	"log"
	"sort"
	"time"

	"github.com/issue9/web/app"
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

func (ms *Modules) newModule(name, desc string, deps ...string) *Module {
	return &Module{
		Name:        name,
		Description: desc,
		Deps:        deps,
		ms:          ms,
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
func (ms *Modules) NewModule(name, desc string, deps ...string) *Module {
	m := ms.newModule(name, desc, deps...)
	ms.modules = append(ms.modules, m)
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

// Init 初始化所有的模块或是模块下指定标签名称的函数。
//
// 若指定了 tag 参数，则只初始化该名称的子模块内容。
func (ms *Modules) Init(tag string, info *log.Logger) error {
	info.Println("开始初始化模块...")

	if err := newDepencency(ms.modules, info).init(tag); err != nil {
		return err
	}

	if tag == "" { // 只有模块的初始化才带路由
		all := ms.app.Mux().All(true, true)
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

// Modules 当前系统使用的所有模块信息
func (ms *Modules) Modules() []*Module {
	return ms.modules
}

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddCron(title string, f app.JobFunc, spec string, delay bool) {
	m.AddInit(func() error {
		return m.ms.app.Scheduled().NewCron(title, f, spec, delay)
	}, "注册计划任务"+title)
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddTicker(title string, f app.JobFunc, dur time.Duration, delay bool) {
	m.AddInit(func() error {
		return m.ms.app.Scheduled().NewTicker(title, f, dur, delay)
	}, "注册计划任务"+title)
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddAt(title string, f app.JobFunc, spec string, delay bool) {
	m.AddInit(func() error {
		return m.ms.app.Scheduled().NewAt(title, f, spec, delay)
	}, "注册计划任务"+title)
}

// AddService 添加新的服务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明。
func (m *Module) AddService(f app.ServiceFunc, title string) {
	m.AddInit(func() error {
		m.ms.app.AddService(f, title)
		return nil
	}, "注册服务："+title)
}
