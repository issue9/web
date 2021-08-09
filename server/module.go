// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"plugin"
	"sort"
	"time"

	"github.com/issue9/scheduled"
	"github.com/issue9/scheduled/schedulers"
	"github.com/issue9/sliceutil"
	"github.com/issue9/web/service"
)

// PluginInitFuncName 插件中的用于获取模块信息的函数名
//
// NOTE: 必须为可导出的函数名称
const PluginInitFuncName = "InitModule"

// PluginInitFunc 安装插件的函数签名
type PluginInitFunc func(*Server) error

// Module 用于注册初始化模块的相关功能
type Module struct {
	tags    map[string]*Tag
	id      string
	desc    string
	version string
	deps    []string

	srv *Server
	fs.FS
}

// Tag 模块下对执行函数的分类
type Tag struct {
	name      string
	m         *Module
	inited    bool
	executors []executor // 保证按添加顺序执行
}

type executor struct {
	title string
	f     func() error
}

// NewModule 声明一个新的模块
//
// id 模块名称，需要全局唯一；
// version 模块的版本信息；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func (srv *Server) NewModule(id, version, desc string, deps ...string) (*Module, error) {
	if sliceutil.Count(srv.modules, func(i int) bool { return srv.modules[i].id == id }) > 0 {
		return nil, fmt.Errorf("存在同名的模块 %s", id)
	}

	sub, err := fs.Sub(srv.fs, id)
	if err != nil {
		return nil, err
	}

	mod := &Module{
		tags:    make(map[string]*Tag, 2),
		id:      id,
		version: version,
		desc:    desc,
		deps:    deps,

		srv: srv,
		FS:  sub,
	}

	srv.modules = append(srv.modules, mod)
	return mod, nil
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
// 如果存在，会调用该方法将插件加载到当前对象中，否则返回相应的错误信息。
func (srv *Server) LoadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return err
	}

	symbol, err := p.Lookup(PluginInitFuncName)
	if err != nil {
		return err
	}

	if install, ok := symbol.(func(*Server) error); ok {
		return install(srv)
	}
	return fmt.Errorf("插件 %s 未找到安装函数", path)
}

// Tag 返回指定名称的 Tag 实例
//
// 如果不存在则会创建。
func (m *Module) Tag(t string) *Tag {
	ev, found := m.tags[t]
	if !found {
		ev = &Tag{
			name:      t,
			executors: make([]executor, 0, 5),
			m:         m,
		}
		m.tags[t] = ev
	}
	return ev
}

// Modules 模块列表
func (srv *Server) Modules() []*Module { return srv.modules }

// ID 模块的唯一 ID
func (m *Module) ID() string { return m.id }

// Description 对模块的详细描述
func (m *Module) Description() string { return m.desc }

// Deps 模块的依赖信息
func (m *Module) Deps() []string { return m.deps }

// Version 版本号
func (m *Module) Version() string { return m.version }

// Tags 模块的标签名称列表
func (m *Module) Tags() []string {
	tags := make([]string, 0, len(m.tags))
	for name := range m.tags {
		tags = append(tags, name)
	}
	sort.Strings(tags)
	return tags
}

// Inited 查询指定标签关联的函数是否已经被调用
func (m *Module) Inited(tag string) bool { return m.Tag(tag).Inited() }

// AddInit 注册指执行函数
//
// NOTE: 按添加顺序执行各个函数。
func (t *Tag) AddInit(title string, f func() error) *Tag {
	t.executors = append(t.executors, executor{title: title, f: f})
	return t
}

func (t *Tag) init(l *log.Logger) error {
	const indent = "\t"

	if t.Inited() {
		return nil
	}

	for _, exec := range t.executors {
		l.Printf("%s%s......", indent, exec.title)
		if err := exec.f(); err != nil {
			l.Printf("%s%s FAIL: %s\n", indent, exec.title, err.Error())
			return err
		}
		l.Printf("%s%s OK", indent, exec.title)
	}

	t.inited = true
	return nil
}

// Inited 当前标签关联的函数是否已经执行过
func (t *Tag) Inited() bool { return t.inited }

// Module 返回当前关联的模块
func (t *Tag) Module() *Module { return t.m }

// AddService 添加新的服务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明。
func (t *Tag) AddService(title string, f service.Func) *Tag {
	return t.AddInit("注册服务："+title, func() error {
		t.Server().Services().AddService(title, f)
		return nil
	})
}

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (t *Tag) AddCron(title string, f scheduled.JobFunc, spec string, delay bool) *Tag {
	return t.AddInit("注册计划任务"+title, func() error {
		return t.Server().Services().AddCron(title, f, spec, delay)
	})
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// imm 是否立即执行一次该任务；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (t *Tag) AddTicker(title string, f scheduled.JobFunc, dur time.Duration, imm, delay bool) *Tag {
	return t.AddInit("注册计划任务"+title, func() error {
		return t.Server().Services().AddTicker(title, f, dur, imm, delay)
	})
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// t 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (t *Tag) AddAt(title string, f scheduled.JobFunc, ti time.Time, delay bool) *Tag {
	return t.AddInit("注册计划任务"+title, func() error {
		return t.Server().Services().AddAt(title, f, ti, delay)
	})
}

// AddJob 添加新的计划任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// scheduler 计划任务的时间调度算法实现；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (t *Tag) AddJob(title string, f scheduled.JobFunc, scheduler schedulers.Scheduler, delay bool) *Tag {
	return t.AddInit("注册计划任务"+title, func() error {
		t.Server().Services().AddJob(title, f, scheduler, delay)
		return nil
	})
}

func (t *Tag) Name() string { return t.name }
