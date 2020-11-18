// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"plugin"
	"sort"
	"strings"
	"time"

	"github.com/issue9/scheduled"
	"github.com/issue9/scheduled/schedulers"

	"github.com/issue9/web/service"
)

// 插件中的初始化函数名称，必须为可导出的函数名称
const moduleInstallFuncName = "Init"

// ErrInited 当模块被多次初始化时返回此错误
var ErrInited = errors.New("模块已经初始化")

type (
	// InstallFunc 安装模块的函数签名
	InstallFunc func(*Server)

	// Module 表示模块信息
	//
	// 模块仅作为在初始化时在代码上的一种分类，一旦初始化完成，
	// 则不再有模块的概念，修改模块的相关属性，也不会对代码有实质性的改变。
	Module struct {
		srv *Server

		Name        string
		Description string
		Deps        []string
		tags        map[string]*Tag
		filters     []Filter
		inits       []*initialization
		inited      bool
	}

	// Tag 表示与特定标签相关联的初始化函数列表
	//
	// 依附于模块，共享模块的依赖关系。
	//
	// 一般是各个模块下的安装脚本使用。
	Tag struct {
		m     *Module
		inits []*initialization
	}

	// 表示初始化功能的相关数据
	initialization struct {
		title string
		f     func() error
	}
)

// NewModule 声明一个新的模块
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func (srv *Server) NewModule(name, desc string, deps ...string) *Module {
	m := &Module{
		srv:         srv,
		Name:        name,
		Description: desc,
		Deps:        deps,
	}
	srv.modules = append(srv.modules, m)
	return m
}

// AddInit 添加一个初始化函数
//
// title 该初始化函数的名称。
func (t *Tag) AddInit(f func() error, title string) {
	if t.m.inited {
		panic(ErrInited)
	}

	if t.inits == nil {
		t.inits = make([]*initialization, 0, 5)
	}
	t.inits = append(t.inits, &initialization{f: f, title: title})
}

// Tags 返回所有的子模块名称
//
// 键名为模块名称，键值为该模块下的标签列表。
func (srv *Server) Tags() map[string][]string {
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
func (srv *Server) Modules() []*Module {
	return srv.modules
}

// InitTag 初始化模块下的子标签
//
// info 用于打印初始化过程的一些信息，如果为空，则采用 web.logs.INFO()。
func (srv *Server) InitTag(tag string, info *log.Logger) error {
	if tag == "" {
		panic("tag 不能为空")
	}

	return srv.init("", srv.Logs().INFO())
}

func (srv *Server) init(tag string, info *log.Logger) error {
	if srv.inited && tag == "" {
		return ErrInited
	}

	if info == nil {
		info = srv.Logs().INFO()
	}

	info.Println("开始初始化模块...")

	if err := srv.initDeps(tag, info); err != nil {
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

	if tag == "" {
		srv.inited = true
	}

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

	if install, ok := symbol.(func(*Server)); ok {
		InstallFunc(install)(srv)
		return nil
	}
	return fmt.Errorf("插件 %s 未找到安装函数", path)
}

// AddService 添加新的服务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明。
func (m *Module) AddService(f service.Func, title string) {
	m.AddInit(func() error {
		m.srv.Services().AddService(f, title)
		return nil
	}, "注册服务："+title)
}

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddCron(title string, f scheduled.JobFunc, spec string, delay bool) {
	m.AddInit(func() error {
		return m.srv.Services().AddCron(title, f, spec, delay)
	}, "注册计划任务"+title)
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// imm 是否立即执行一次该任务；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddTicker(title string, f scheduled.JobFunc, dur time.Duration, imm, delay bool) {
	m.AddInit(func() error {
		return m.srv.Services().AddTicker(title, f, dur, imm, delay)
	}, "注册计划任务"+title)
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// t 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddAt(title string, f scheduled.JobFunc, t time.Time, delay bool) {
	m.AddInit(func() error {
		return m.srv.Services().AddAt(title, f, t, delay)
	}, "注册计划任务"+title)
}

// AddJob 添加新的计划任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// scheduler 计划任务的时间调度算法实现；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddJob(title string, f scheduled.JobFunc, scheduler schedulers.Scheduler, delay bool) {
	m.AddInit(func() error {
		m.srv.Services().AddJob(title, f, scheduler, delay)
		return nil
	}, "注册计划任务"+title)
}

// AddInit 添加一个初始化函数
//
// title 该初始化函数的名称。
func (m *Module) AddInit(f func() error, title string) {
	if m.inited {
		panic(ErrInited)
	}

	if m.inits == nil {
		m.inits = make([]*initialization, 0, 5)
	}

	m.inits = append(m.inits, &initialization{f: f, title: title})
}

// NewTag 为当前模块生成特定名称的子模块
//
// 若已经存在，则直接返回该子模块。
//
// Tag 是依赖关系与当前模块相同，但是功能完全独立的模块，
// 一般用于功能更新等操作。
func (m *Module) NewTag(tag string) *Tag {
	if m.tags == nil {
		m.tags = make(map[string]*Tag, 5)
	}

	if _, found := m.tags[tag]; !found {
		m.tags[tag] = &Tag{
			m:     m,
			inits: make([]*initialization, 0, 5),
		}
	}

	return m.tags[tag]
}
