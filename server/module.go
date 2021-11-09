// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"io/fs"
	"log"
	"sort"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/message"

	"github.com/issue9/web/internal/dep"
	"github.com/issue9/web/internal/filesystem"
)

type (
	// PluginInitFunc 安装插件的函数签名
	PluginInitFunc func(*Server) error

	// Module 用于注册初始化模块的相关功能
	Module struct {
		actions map[string]*Action
		id      string
		desc    localeutil.LocaleStringer
		version string
		deps    []string

		inits   []dep.Executor
		uninits []dep.Executor

		srv    *Server
		fs     *filesystem.MultipleFS
		object interface{}
	}

	ModuleInfo struct {
		ID          string   `yaml:"id" json:"id" xml:"id,attr"`
		Version     string   `yaml:"version" json:"version" xml:"version,attr"`
		Description string   `yaml:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
		Deps        []string `yaml:"deps,omitempty" json:"deps,omitempty" xml:"dep,omitempty"`
	}

	// Action 模块下对初始化函数的分组
	Action struct {
		name string
		m    *Module

		inits   []dep.Executor // 保证按添加顺序执行
		uninits []dep.Executor
	}
)

func (srv *Server) initModules(uninit bool, action string) error {
	if action == "" {
		panic("参数  action 不能为空")
	}

	items := make([]*dep.Item, 0, len(srv.modules))
	if uninit {
		for _, m := range srv.modules {
			executors := append(m.uninits, m.Action(action).uninits...)
			items = append(items, dep.NewItem(m.id, m.deps, executors))
		}
		items = dep.Reverse(items)
	} else {
		for _, m := range srv.modules {
			executors := append(m.inits, m.Action(action).inits...)
			items = append(items, dep.NewItem(m.id, m.deps, executors))
		}
	}

	// 日志不需要标出文件位置
	l := srv.Logs().INFO()
	flags := l.Flags()
	l.SetFlags(log.Ldate | log.Lmicroseconds)
	defer l.SetFlags(flags)

	p := srv.LocalePrinter()
	l.Println(p.Sprintf("run action %s...", action))
	if err := dep.Dep(l, items); err != nil {
		return err
	}
	l.Println(p.Sprintf("action %s complete", action))

	return nil
}

// NewModule 声明一个新的模块
//
// id 模块名称，需要全局唯一；
// version 模块的版本信息；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func (srv *Server) NewModule(id, version string, desc localeutil.LocaleStringer, deps ...string) *Module {
	if sliceutil.Count(srv.modules, func(i int) bool { return srv.modules[i].id == id }) > 0 {
		panic(fmt.Sprintf("存在同名的模块 %s", id))
	}

	sub, err := fs.Sub(srv.fs, id)
	if err != nil {
		panic(err)
	}

	mod := &Module{
		actions: make(map[string]*Action, 2),
		id:      id,
		version: version,
		desc:    desc,
		deps:    deps,

		inits:   make([]dep.Executor, 0, 3),
		uninits: make([]dep.Executor, 0, 3),

		srv: srv,
		fs:  filesystem.NewMultipleFS(sub),
	}

	srv.modules = append(srv.modules, mod)
	return mod
}

// Actions 返回 Action 列表
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

// Actions 返回 Action 列表
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

// DepObject 获取依赖项关联的对象
func (m *Module) DepObject(depID string) interface{} {
	// TODO: 每次都需要构建 items，想办法简化掉！
	items := make([]*dep.Item, 0, len(m.Server().modules))
	for _, m := range m.Server().modules {
		items = append(items, dep.NewItem(m.id, m.deps, nil))
	}

	if !dep.IsDep(items, m.id, depID) {
		panic(fmt.Sprintf("%s 并不是 %s 的依赖对象", depID, m.id))
	}

	for _, mod := range m.srv.modules {
		if mod.id == depID {
			return mod.object
		}
	}
	panic(fmt.Sprintf("依赖项 %s 未找到", depID))
}

// AddInit 注册模块初始化时执行的函数
func (m *Module) AddInit(title string, f func() error) {
	m.inits = append(m.inits, dep.Executor{Title: title, F: f})
}

// AddUninit 注册卸载模块时的执行函数
func (m *Module) AddUninit(title string, f func() error) {
	m.uninits = append(m.uninits, dep.Executor{Title: title, F: f})
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

// AddInit 注册模块初始化时执行的函数
//
// NOTE: 按添加顺序执行各个函数。
func (t *Action) AddInit(title string, f func() error) *Action {
	t.inits = append(t.inits, dep.Executor{Title: title, F: f})
	return t
}

// AddUninit 注册卸载模块时的执行函数
//
// NOTE: 按添加顺序执行各个函数。
func (t *Action) AddUninit(title string, f func() error) *Action {
	t.uninits = append(t.uninits, dep.Executor{Title: title, F: f})
	return t
}

// Module 返回当前关联的模块
func (t *Action) Module() *Module { return t.m }

func (t *Action) Name() string { return t.name }
