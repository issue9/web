// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"log"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	"github.com/issue9/scheduled"
	"github.com/issue9/scheduled/schedulers"

	"github.com/issue9/web/module"
	"github.com/issue9/web/service"
)

// ModuleFuncName 插件中的用于获取模块信息的函数名
//
// NOTE: 必须为可导出的函数名称
const ModuleFuncName = "Module"

// ModuleFunc 安装插件的函数签名
type ModuleFunc func(*Server) (*Module, error)

// Module 表示模块信息
//
// 模块可以作为代码的一种组织方式。将一组关联的功能合并为一个模块。
type Module struct {
	*module.Module
	srv     *Server
	filters []Filter
}

// NewModule 声明一个新的模块
//
// id 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func (srv *Server) NewModule(id, desc string, deps ...string) *Module {
	return &Module{
		Module: module.NewModule(id, desc, deps...),
		srv:    srv,
	}
}

// AddModuleFunc 从 ModuleFunc 添加模块
func (srv *Server) AddModuleFunc(module ...ModuleFunc) error {
	ms := make([]*Module, 0, len(module))
	for _, f := range module {
		m, err := f(srv)
		if err != nil {
			return err
		}
		ms = append(ms, m)
	}

	return srv.AddModule(ms...)
}

// AddModule 添加模块
//
// 可以在运行过程中添加模块，该模块会在加载时直接初始化，前提是模块的依赖模块都已经初始化。
func (srv *Server) AddModule(module ...*Module) error {
	for _, m := range module {
		m.srv = srv

		if err := srv.dep.Add(m.Module); err != nil {
			return err
		}
	}
	return nil
}

// Tags 返回所有的子模块名称
//
// 键名为模块名称，键值为该模块下的标签列表。
func (srv *Server) Tags() []string { return srv.dep.Tags() }

// Modules 当前系统使用的所有模块信息
func (srv *Server) Modules() []*module.Module { return srv.dep.Modules() }

// InitTag 初始化模块下的子标签
func (srv *Server) InitTag(tag string) error {
	if tag == "" {
		panic("参数  tag 不能为空")
	}

	return srv.dep.Init(tag)
}

// initModules 初始化模块
func (srv *Server) initModules() error {
	// 以下输出日志不需要标出文件位置。
	info := srv.Logs().INFO()
	flags := info.Flags()
	info.SetFlags(log.Ldate | log.Lmicroseconds)

	info.Println("开始初始化模块...")

	if err := srv.dep.Init(""); err != nil {
		return err
	}

	routers := srv.Mux().Routers()
	for _, r := range routers {
		if routes := r.Routes(true, true); len(routes) > 0 {
			info.Printf("模块加载了 %s 的以下路由项：\n", r.Name())
			for path, methods := range routes {
				info.Printf("\t[%s] %s\n", strings.Join(methods, ", "), path)
			}
		}
	}

	info.Println("模块初始化完成！")

	info.SetFlags(flags)

	return nil
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

	if install, ok := symbol.(func(*Server) (*Module, error)); ok {
		return srv.AddModuleFunc(install)
	}

	return fmt.Errorf("插件 %s 未找到安装函数", path)
}

// AddService 添加新的服务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明。
func (m *Module) AddService(title string, f service.Func) {
	m.AddInit("注册服务："+title, func() error {
		m.srv.Services().AddService(f, title)
		return nil
	})
}

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddCron(title string, f scheduled.JobFunc, spec string, delay bool) {
	m.AddInit("注册计划任务"+title, func() error {
		return m.srv.Services().AddCron(title, f, spec, delay)
	})
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// imm 是否立即执行一次该任务；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddTicker(title string, f scheduled.JobFunc, dur time.Duration, imm, delay bool) {
	m.AddInit("注册计划任务"+title, func() error {
		return m.srv.Services().AddTicker(title, f, dur, imm, delay)
	})
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// t 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddAt(title string, f scheduled.JobFunc, t time.Time, delay bool) {
	m.AddInit("注册计划任务"+title, func() error {
		return m.srv.Services().AddAt(title, f, t, delay)
	})
}

// AddJob 添加新的计划任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// scheduler 计划任务的时间调度算法实现；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddJob(title string, f scheduled.JobFunc, scheduler schedulers.Scheduler, delay bool) {
	m.AddInit("注册计划任务"+title, func() error {
		m.srv.Services().AddJob(title, f, scheduler, delay)
		return nil
	})
}

// NewTag 为当前模块生成特定名称的子模块
//
// 若已经存在，则直接返回该子模块。
//
// Tag 是依赖关系与当前模块相同，但是功能完全独立的模块，
// 一般用于功能更新等操作。
func (m *Module) NewTag(tag string) module.Initializer { return m.GetTag(tag) }
