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
)

// 插件中的初始化函数名称，必须为可导出的函数名称
const moduleInstallFuncName = "Init"

// ErrInited 当模块被多次初始化时返回此错误
var ErrInited = errors.New("模块已经初始化")

type (
	// InstallFunc 安装模块的函数签名
	InstallFunc func(*Web)

	// Module 表示模块信息
	//
	// 模块仅作为在初始化时在代码上的一种分类，一旦初始化完成，
	// 则不再有模块的概念，修改模块的相关属性，也不会对代码有实质性的改变。
	Module struct {
		Name        string
		Description string
		Deps        []string
		tags        map[string]*Tag
		web         *Web
		filters     []Filter
		inits       []*initialization

		inited bool
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

// NewModule 声明一个新的模块
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func (web *Web) NewModule(name, desc string, deps ...string) *Module {
	m := &Module{
		Name:        name,
		Description: desc,
		Deps:        deps,
		web:         web,
	}
	web.modules = append(web.modules, m)
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
func (web *Web) Tags() map[string][]string {
	ret := make(map[string][]string, len(web.modules)*2)

	for _, m := range web.modules {
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
func (web *Web) Modules() []*Module {
	return web.modules
}

// Init 初始化模块
//
// 若指定了 tag 参数，则只初始化与该标签相关联的内容；
// info 用于打印初始化过程的一些信息，如果为空，则采用 web.logs.INFO()。
//
// 一旦初始化完成，则不再接受添加新模块，也不能再次进行初始化。
// Web 的大部分功能将失去操作意义，比如 Web.NewModule
// 虽然能添加新模块到 Server，但并不能真正初始化新的模块并挂载。
func (web *Web) Init(tag string, info *log.Logger) error {
	if web.inited && tag == "" {
		return ErrInited
	}

	if info == nil {
		info = web.Logs().INFO()
	}

	info.Println("开始初始化模块...")

	if err := web.initDeps(tag, info); err != nil {
		return err
	}

	if all := web.ctxServer.Router().Mux().All(true, true); len(all) > 0 {
		info.Println("模块加载了以下路由项：")
		for path, methods := range all {
			info.Printf("[%s] %s\n", strings.Join(methods, ", "), path)
		}
	}

	info.Println("模块初始化完成！")

	if tag == "" {
		web.inited = true
	}

	return nil
}

// 加载所有的插件
//
// 如果 glob 为空，则不会加载任何内容，返回空值
func (web *Web) loadPlugins(glob string) error {
	fs, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	for _, path := range fs {
		if err := web.loadPlugin(path); err != nil {
			return err
		}
	}

	return nil
}

func (web *Web) loadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return err
	}

	symbol, err := p.Lookup(moduleInstallFuncName)
	if err != nil {
		return err
	}

	if install, ok := symbol.(func(*Web)); ok {
		InstallFunc(install)(web)
		return nil
	}
	return fmt.Errorf("插件 %s 未找到安装函数", path)
}
