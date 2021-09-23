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

	"github.com/issue9/localeutil"
	"github.com/issue9/scheduled"
	"github.com/issue9/scheduled/schedulers"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/message"

	"github.com/issue9/web/internal/dep"
	"github.com/issue9/web/internal/filesystem"
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
	actions map[string]*Action
	id      string
	desc    localeutil.LocaleStringer
	version string
	deps    []string

	srv    *Server
	fs     *filesystem.MultipleFS
	object interface{}
}

type ModuleInfo struct {
	ID          string   `yaml:"id" json:"id" xml:"id,attr"`
	Version     string   `yaml:"version" json:"version" xml:"version,attr"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
	Deps        []string `yaml:"deps,omitempty" json:"deps,omitempty" xml:"dep,omitempty"`
}

// Action 模块下对初始化函数的分组
type Action struct {
	name   string
	m      *Module
	inited bool

	inits   []dep.Executor // 保证按添加顺序执行
	uninits []dep.Executor
}

func (srv *Server) initModules(action string, uninit bool) error {
	if action == "" {
		panic("参数  action 不能为空")
	}

	items := make([]*dep.Item, 0, len(srv.modules))
	if uninit {
		for _, m := range srv.modules {
			items = append(items, &dep.Item{
				ID:        m.id,
				Deps:      m.deps,
				Executors: m.Action(action).uninits,
			})
		}
		items = dep.Reverse(items)
	} else {
		for _, m := range srv.modules {
			items = append(items, &dep.Item{
				ID:        m.id,
				Deps:      m.deps,
				Executors: m.Action(action).inits,
			})
		}
	}

	// 日志不需要标出文件位置
	l := srv.Logs().INFO()
	flags := l.Flags()
	l.SetFlags(log.Ldate | log.Lmicroseconds)
	defer l.SetFlags(flags)

	l.Printf("开始初始化模块中的 %s...\n", action)
	if err := dep.Dep(l, items); err != nil {
		return err
	}
	l.Print("初始化完成！\n\n")

	return nil
}

// NewModule 声明一个新的模块
//
// id 模块名称，需要全局唯一；
// version 模块的版本信息；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func (srv *Server) NewModule(id, version string, desc localeutil.LocaleStringer, deps ...string) (*Module, error) {
	if sliceutil.Count(srv.modules, func(i int) bool { return srv.modules[i].id == id }) > 0 {
		return nil, fmt.Errorf("存在同名的模块 %s", id)
	}

	sub, err := fs.Sub(srv.fs, id)
	if err != nil {
		return nil, err
	}

	mod := &Module{
		actions: make(map[string]*Action, 2),
		id:      id,
		version: version,
		desc:    desc,
		deps:    deps,

		srv: srv,
		fs:  filesystem.NewMultipleFS(sub),
	}

	srv.modules = append(srv.modules, mod)
	return mod, nil
}

// loadPlugins 加载所有的插件
//
// 如果 glob 为空，则不会加载任何内容，返回空值。
func (srv *Server) loadPlugins(glob string) error {
	fsys, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	for _, path := range fsys {
		if err := srv.loadPlugin(path); err != nil {
			return err
		}
	}

	return nil
}

// loadPlugin 将指定的插件当作模块进行加载
//
// path 为插件的路径；
//
// 插件必须是以 buildmode=plugin 的方式编译的，且要求其引用的 github.com/issue9/web
// 版本与当前的相同。
// loadPlugin 会在插件中查找固定名称和类型的函数名（参考 PluginInitFunc 和 PluginInitFuncName），
// 如果存在，会调用该方法将插件加载到当前对象中，否则返回相应的错误信息。
func (srv *Server) loadPlugin(path string) error {
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

func (srv *Server) Actions() []string {
	actions := make([]string, 0, 100)
	for _, m := range srv.modules {
		actions = append(actions, m.Actions()...)
	}
	size := sliceutil.Unique(actions, func(i, j int) bool { return actions[i] == actions[j] })
	actions = actions[:size]
	sort.Strings(actions)
	return actions
}

// Action 返回指定名称的 Action 实例
//
// 如果不存在则会创建。
func (m *Module) Action(t string) *Action {
	ev, found := m.actions[t]
	if !found {
		ev = &Action{
			name:    t,
			inits:   make([]dep.Executor, 0, 5),
			uninits: make([]dep.Executor, 0, 5),
			m:       m,
		}
		m.actions[t] = ev
	}
	return ev
}

// Modules 模块列表
func (srv *Server) Modules(p *message.Printer) []*ModuleInfo {
	info := make([]*ModuleInfo, 0, len(srv.modules))
	for _, m := range srv.modules {
		info = append(info, &ModuleInfo{
			ID:          m.id,
			Version:     m.version,
			Description: m.desc.LocaleString(p),
			Deps:        m.deps,
		})
	}
	return info
}

func (m *Module) Open(name string) (fs.File, error) { return m.fs.Open(name) }

func (m *Module) Actions() []string {
	actions := make([]string, 0, len(m.actions))
	for name := range m.actions {
		actions = append(actions, name)
	}
	sort.Strings(actions)
	return actions
}

// AttachObject 为当前模块附加一个对象
//
// 若模块中有需要外放的数据，可以通过此方法将数据附加在模块上。
func (m *Module) AttachObject(v interface{}) { m.object = v }

// Object 获取通过 AttachObject 关联的对象
func (m *Module) Object() interface{} { return m.object }

// DepObject 获取依赖项关联的对象
func (m *Module) DepObject(dep string) interface{} {
	if sliceutil.Index(m.deps, func(i int) bool { return m.deps[i] == dep }) < 0 {
		panic(fmt.Sprintf("%s 并不是 %s 的依赖对象", dep, m.id))
	}

	var obj *Module
	for _, mod := range m.srv.modules {
		if mod.id == dep {
			obj = mod
			break
		}
	}
	if obj == nil {
		panic(fmt.Sprintf("依赖项 %s 未找到", dep))
	}
	return obj.Object()
}

// LoadLocale 从 m.FS 加载本地化语言文件
func (m *Module) LoadLocale(glob string) error {
	return m.srv.Locale().LoadFileFS(m, glob)
}

// AddFS 将多个文件系统与当前模块的文件系统进行关联
//
// 当采用 Module.Open 查找文件时，会根据添加的顺序依次查找文件，
// 只要存在于某一个文件系统中，那么就当作该文件存在，并返回。
//
// 每个模块在初始化时，都会默认将 Server.FS + Module.ID
// 作为模块的文件系统，通过 AddFS 可以挂载其它的文件系统，
// 与 embed.FS 相结合，可以做到在外部相对应目录中有修改时，
// 读取外部的文件，如果不存在，则读取 embed.FS 中的内容。
func (m *Module) AddFS(fsys ...fs.FS) { m.fs.Add(fsys...) }

// AddInit 注册指执行函数
//
// NOTE: 按添加顺序执行各个函数。
func (t *Action) AddInit(title string, f func() error) *Action {
	t.inits = append(t.inits, dep.Executor{Title: title, F: f})
	return t
}

// AddUninit 注册指执行函数
//
// NOTE: 按添加顺序执行各个函数。
func (t *Action) AddUninit(title string, f func() error) *Action {
	t.uninits = append(t.uninits, dep.Executor{Title: title, F: f})
	return t
}

// Module 返回当前关联的模块
func (t *Action) Module() *Module { return t.m }

// AddService 添加新的服务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明。
func (t *Action) AddService(title string, f service.Func) *Action {
	msg := t.Server().LocalePrinter().Sprintf("register service", title)
	return t.AddInit(msg, func() error {
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
func (t *Action) AddCron(title string, f scheduled.JobFunc, spec string, delay bool) *Action {
	msg := t.Server().LocalePrinter().Sprintf("register cron", title)
	return t.AddInit(msg, func() error {
		return t.Server().Services().AddCron(title, f, spec, delay)
	})
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// imm 是否立即执行一次该任务；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (t *Action) AddTicker(title string, f scheduled.JobFunc, dur time.Duration, imm, delay bool) *Action {
	msg := t.Server().LocalePrinter().Sprintf("register cron", title)
	return t.AddInit(msg, func() error {
		return t.Server().Services().AddTicker(title, f, dur, imm, delay)
	})
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// t 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (t *Action) AddAt(title string, f scheduled.JobFunc, ti time.Time, delay bool) *Action {
	msg := t.Server().LocalePrinter().Sprintf("register cron", title)
	return t.AddInit(msg, func() error {
		return t.Server().Services().AddAt(title, f, ti, delay)
	})
}

// AddJob 添加新的计划任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// scheduler 计划任务的时间调度算法实现；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (t *Action) AddJob(title string, f scheduled.JobFunc, scheduler schedulers.Scheduler, delay bool) *Action {
	msg := t.Server().LocalePrinter().Sprintf("register cron", title)
	return t.AddInit(msg, func() error {
		t.Server().Services().AddJob(title, f, scheduler, delay)
		return nil
	})
}

func (t *Action) Name() string { return t.name }
