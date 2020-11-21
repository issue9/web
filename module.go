// SPDX-License-Identifier: MIT

package web

import (
	"fmt"
	"log"
	"path/filepath"
	"plugin"
	"sort"
	"strings"
	"time"

	"github.com/issue9/scheduled"
	"github.com/issue9/scheduled/schedulers"

	"github.com/issue9/web/internal/dep"
	"github.com/issue9/web/service"
)

// 插件中的初始化函数名称，必须为可导出的函数名称
const moduleInstallFuncName = "Init"

type (
	// InstallFunc 安装模块的函数签名
	InstallFunc func(*Server) error

	// Module 表示模块信息
	//
	// 模块仅作为在初始化时在代码上的一种分类，一旦初始化完成，
	// 则不再有模块的概念，修改模块的相关属性，也不会对代码有实质性的改变。
	Module interface {
		ID() string
		Description() string
		Deps() []string

		Tags() []string // 与当前模块关联的子标签

		// 添加新的服务
		//
		// f 表示服务的运行函数；
		// title 是对该服务的简要说明。
		AddService(title string, f service.Func)

		// AddCron 添加新的定时任务
		//
		// f 表示服务的运行函数；
		// title 是对该服务的简要说明；
		// spec cron 表达式，支持秒；
		// delay 是否在任务执行完之后，才计算下一次的执行时间点。
		AddCron(title string, f scheduled.JobFunc, spec string, delay bool)

		// AddTicker 添加新的定时任务
		//
		// f 表示服务的运行函数；
		// title 是对该服务的简要说明；
		// imm 是否立即执行一次该任务；
		// delay 是否在任务执行完之后，才计算下一次的执行时间点。
		AddTicker(title string, f scheduled.JobFunc, dur time.Duration, imm, delay bool)

		// AddAt 添加新的定时任务
		//
		// f 表示服务的运行函数；
		// title 是对该服务的简要说明；
		// t 指定的时间点；
		// delay 是否在任务执行完之后，才计算下一次的执行时间点。
		AddAt(title string, f scheduled.JobFunc, t time.Time, delay bool)

		// AddJob 添加新的计划任务
		//
		// f 表示服务的运行函数；
		// title 是对该服务的简要说明；
		// scheduler 计划任务的时间调度算法实现；
		// delay 是否在任务执行完之后，才计算下一次的执行时间点。
		AddJob(title string, f scheduled.JobFunc, scheduler schedulers.Scheduler, delay bool)

		// AddInit 添加一个初始化函数
		//
		// title 该初始化函数的名称。
		AddInit(title string, f func() error)

		// NewTag 为当前模块生成特定名称的子模块
		//
		// 若已经存在，则直接返回该子模块。
		//
		// Tag 是依赖关系与当前模块相同，但是功能完全独立的模块，
		// 一般用于功能更新等操作。
		NewTag(tag string) (Tag, error)

		// AddFilters 添加过滤器
		//
		// 按给定参数的顺序反向依次调用。
		AddFilters(filter ...Filter) Module
		Resource(pattern string, filter ...Filter) Resource
		Prefix(prefix string, filter ...Filter) Prefix
		Handle(path string, h HandlerFunc, method ...string) Module
		Get(path string, h HandlerFunc) Module
		Post(path string, h HandlerFunc) Module
		Delete(path string, h HandlerFunc) Module
		Put(path string, h HandlerFunc) Module
		Patch(path string, h HandlerFunc) Module
		Options(path, allow string) Module
		Remove(path string, method ...string) Module
	}

	// Tag 表示与特定标签相关联的初始化函数列表
	//
	// 依附于模块，共享模块的依赖关系。
	// 一般是各个模块下的安装脚本使用。
	Tag interface {
		AddInit(string, func() error)
	}

	mod struct {
		*dep.Default
		srv     *Server
		filters []Filter
	}
)

// NewModule 声明一个新的模块
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func (srv *Server) NewModule(id, desc string, deps ...string) (Module, error) {
	m := &mod{
		Default: dep.NewDefaultModule(id, desc, deps...),
		srv:     srv,
	}

	if err := srv.modules.AddModule(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Tags 返回所有的子模块名称
//
// 键名为模块名称，键值为该模块下的标签列表。
func (srv *Server) Tags() map[string][]string {
	ret := make(map[string][]string, len(srv.tags))

	for name, d := range srv.tags {
		for _, tag := range d.Modules() {
			ret[tag.ID()] = append(ret[tag.ID()], name)
			sort.Strings(ret[tag.ID()])
		}
	}

	return ret
}

// Modules 当前系统使用的所有模块信息
func (srv *Server) Modules() []dep.Module {
	return srv.modules.Modules()
}

// InitTag 初始化模块下的子标签
func (srv *Server) InitTag(tag string, info *log.Logger) error {
	if tag == "" {
		panic("tag 不能为空")
	}

	tags, found := srv.tags[tag]
	if !found {
		return fmt.Errorf("标签 %s 不存在", tag)
	}
	return tags.Init()
}

// InitModules 初始化模块
func (srv *Server) InitModules(info *log.Logger) error {
	info.Println("开始初始化模块...")

	if err := srv.modules.Init(); err != nil {
		return err
	}

	if all := srv.Router().Mux().All(true, true); len(all) > 0 {
		info.Println("模块加载了以下路由项：")
		for _, router := range all {
			info.Println(router.Name)
			for path, methods := range router.Routes {
				info.Printf("    [%s] %s\n", strings.Join(methods, ", "), path)
			}
		}
	}

	info.Println("模块初始化完成！")

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
func (srv *Server) LoadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return err
	}

	symbol, err := p.Lookup(moduleInstallFuncName)
	if err != nil {
		return err
	}

	if install, ok := symbol.(func(*Server) error); ok {
		return InstallFunc(install)(srv)
	}

	// TODO 如果已经 inited，那么直接初始化插件

	return fmt.Errorf("插件 %s 未找到安装函数", path)
}

func (m *mod) AddService(title string, f service.Func) {
	m.AddInit("注册服务："+title, func() error {
		m.srv.Services().AddService(f, title)
		return nil
	})
}

func (m *mod) AddCron(title string, f scheduled.JobFunc, spec string, delay bool) {
	m.AddInit("注册计划任务"+title, func() error {
		return m.srv.Services().AddCron(title, f, spec, delay)
	})
}

func (m *mod) AddTicker(title string, f scheduled.JobFunc, dur time.Duration, imm, delay bool) {
	m.AddInit("注册计划任务"+title, func() error {
		return m.srv.Services().AddTicker(title, f, dur, imm, delay)
	})
}

func (m *mod) AddAt(title string, f scheduled.JobFunc, t time.Time, delay bool) {
	m.AddInit("注册计划任务"+title, func() error {
		return m.srv.Services().AddAt(title, f, t, delay)
	})
}

func (m *mod) AddJob(title string, f scheduled.JobFunc, scheduler schedulers.Scheduler, delay bool) {
	m.AddInit("注册计划任务"+title, func() error {
		m.srv.Services().AddJob(title, f, scheduler, delay)
		return nil
	})
}

func (m *mod) NewTag(tag string) (Tag, error) {
	if m.srv.tags == nil {
		m.srv.tags = make(map[string]*dep.Dep, 5)
	}

	d, found := m.srv.tags[tag]
	if !found {
		d = dep.New(m.srv.Logs().INFO())
		m.srv.tags[tag] = d
	}

	if mod := d.FindModule(m.ID()); mod != nil {
		return mod.(Tag), nil
	}

	t := dep.NewDefaultModule(m.ID(), m.Description(), m.Deps()...)
	if err := d.AddModule(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (m *mod) Tags() []string {
	// TODO
	return nil
}
