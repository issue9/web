// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"errors"
	"net/http"
	"plugin"

	"github.com/issue9/web/dependency"
)

// PluginModuleName 插件中必须提供的变量名称。
// 根据此变量提供的数据进行初始化。
const PluginModuleName = "Module"

// 表示模块的类型。
const (
	ModuleTypeAll    ModuleType = iota
	ModuleTypeModule            // 默认的方式，即和代码一起编译
	ModuleTypePlugin            // 加载以 buildmode=plugin 方式编译的模块
)

// ModuleType 用以指定模块的类型。
type ModuleType int8

// Module 表示模块信息
type Module struct {
	Name        string
	Deps        []string
	Description string
	Routes      []*Route
	Type        ModuleType

	inits        []dependency.InitFunc
	buildHandler BuildHandler
}

// Route 表示模块信息中的路由信息
type Route struct {
	Path    string
	Handler http.Handler
	Methods []string
}

// Prefix 可以将具有统一前缀的路由项集中在一起操作。
//  p := srv.Prefix("/api")
//  p.Get("/users")  // 相当于 srv.Get("/api/users")
//  p.Get("/user/1") // 相当于 srv.Get("/api/user/1")
type Prefix struct {
	module *Module
	prefix string
}

// Modules 获取当前的所有模块信息
func (app *App) Modules() []*Module {
	return app.modules
}

// AddModule 注册一个新的模块。
func (app *App) AddModule(m *Module) *App {
	app.modules = append(app.modules, m)
	return app
}

// 加载配置文件中指定的所有插件
func (app *App) loadPlugins() error {
	for _, p := range app.config.Plugins {
		if err := app.loadPlugin(p); err != nil {
			return err
		}
	}

	return nil
}

func (app *App) loadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return err
	}

	symbol, err := p.Lookup("M")
	if err != nil {
		return err
	}

	module, ok := symbol.(*Module)
	if !ok {
		return errors.New("无法转换成 Module")
	}

	module.Type = ModuleTypePlugin
	app.AddModule(module)

	return nil
}

func (app *App) initDependency() error {
	dep := dependency.New()

	for _, module := range app.modules {
		dep.Add(module.Name, app.getInit(module), module.Deps...)
	}

	return dep.Init()
}

// 当 Module 的内容生成一个 dependency.InitFunc 函数
func (app *App) getInit(m *Module) dependency.InitFunc {
	return func() error {
		for _, init := range m.inits {
			if err := init(); err != nil {
				return err
			}
		}

		for _, r := range m.Routes {
			h := r.Handler
			if m.buildHandler != nil {
				h = m.buildHandler(h)
			}

			if err := app.router.Handle(r.Path, h, r.Methods...); err != nil {
				return err
			}
		}

		return nil
	}
}

// NewPlugin 声明一个新的模块，该模块以插件的形式提供。
//
// 需要将插件添加到配置文件中，才会在启动时，加载到系统中。
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称。
func NewPlugin(name, desc string, deps ...string) *Module {
	m := NewModule(name, desc, deps...)
	m.Type = ModuleTypePlugin
	return m
}

// NewModule 声明一个新的模块
//
// 仅作声明，并不会添加到系统中，需要通过 AddModule 时行添加。
//
// name 模块名称，需要全局唯一；
// desc 模块的详细信息；
// deps 表示当前模块的依赖模块名称，可以是插件中的模块名称。
func NewModule(name, desc string, deps ...string) *Module {
	return &Module{
		Name:        name,
		Deps:        deps,
		Description: desc,
		Type:        ModuleTypeModule,
		Routes:      make([]*Route, 0, 10),
		inits:       make([]dependency.InitFunc, 0, 5),
	}
}

// Prefix 声明一个 Prefix 实例。
func (m *Module) Prefix(prefix string) *Prefix {
	return &Prefix{
		module: m,
		prefix: prefix,
	}
}

// Module 返回关联的 Module 实全
func (p *Prefix) Module() *Module {
	return p.module
}

// SetHandler 为当前模块的所有路由添加的中间件。
//
// 多次调用，最后一次为准。
func (m *Module) SetHandler(f BuildHandler) {
	m.buildHandler = f
}

// AddInit 添加一个初始化函数
func (m *Module) AddInit(f dependency.InitFunc) *Module {
	m.inits = append(m.inits, f)
	return m
}

// Handle 添加一个路由项
func (m *Module) Handle(path string, h http.Handler, methods ...string) *Module {
	m.Routes = append(m.Routes, &Route{
		Methods: methods,
		Path:    path,
		Handler: h,
	})

	return m
}

// Get 指定一个 GET 请求
func (m *Module) Get(path string, h http.Handler) *Module {
	return m.Handle(path, h, http.MethodGet)
}

// Post 指定个 POST 请求处理
func (m *Module) Post(path string, h http.Handler) *Module {
	return m.Handle(path, h, http.MethodPost)
}

// Delete 指定个 Delete 请求处理
func (m *Module) Delete(path string, h http.Handler) *Module {
	return m.Handle(path, h, http.MethodDelete)
}

// Put 指定个 Put 请求处理
func (m *Module) Put(path string, h http.Handler) *Module {
	return m.Handle(path, h, http.MethodPut)
}

// Patch 指定个 Patch 请求处理
func (m *Module) Patch(path string, h http.Handler) *Module {
	return m.Handle(path, h, http.MethodPatch)
}

// HandleFunc 指定一个请求
func (m *Module) HandleFunc(path string, h func(w http.ResponseWriter, r *http.Request), methods ...string) *Module {
	return m.Handle(path, http.HandlerFunc(h), methods...)
}

// GetFunc 指定一个 GET 请求
func (m *Module) GetFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.HandleFunc(path, h, http.MethodGet)
}

// PostFunc 指定一个 Post 请求
func (m *Module) PostFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.HandleFunc(path, h, http.MethodPost)
}

// DeleteFunc 指定一个 Delete 请求
func (m *Module) DeleteFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.HandleFunc(path, h, http.MethodDelete)
}

// PutFunc 指定一个 Put 请求
func (m *Module) PutFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.HandleFunc(path, h, http.MethodPut)
}

// PatchFunc 指定一个 Patch 请求
func (m *Module) PatchFunc(path string, h func(w http.ResponseWriter, r *http.Request)) *Module {
	return m.HandleFunc(path, h, http.MethodPatch)
}

// Handle 添加路由项
func (p *Prefix) Handle(path string, h http.Handler, methods ...string) *Prefix {
	p.module.Handle(p.prefix+path, h, methods...)
	return p
}

// Get 指定一个 GET 请求
func (p *Prefix) Get(path string, h http.Handler) *Prefix {
	return p.Handle(path, h, http.MethodGet)
}

// Post 指定一个 Post 请求
func (p *Prefix) Post(path string, h http.Handler) *Prefix {
	return p.Handle(path, h, http.MethodPost)
}

// Delete 指定一个 Delete 请求
func (p *Prefix) Delete(path string, h http.Handler) *Prefix {
	return p.Handle(path, h, http.MethodDelete)
}

// Put 指定一个 Put 请求
func (p *Prefix) Put(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodPut)
}

// Patch 指定一个 Patch 请求
func (p *Prefix) Patch(pattern string, h http.Handler) *Prefix {
	return p.Handle(pattern, h, http.MethodPatch)
}

// HandleFunc 指定一个请求
func (p *Prefix) HandleFunc(path string, f http.HandlerFunc, methods ...string) *Prefix {
	return p.Handle(path, http.HandlerFunc(f), methods...)
}

// GetFunc 指定一个 Get 请求
func (p *Prefix) GetFunc(path string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(path, f, http.MethodGet)
}

// PutFunc 指定一个 Put 请求
func (p *Prefix) PutFunc(path string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(path, f, http.MethodPut)
}

// PostFunc 指定一个 Post 请求
func (p *Prefix) PostFunc(path string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(path, f, http.MethodPost)
}

// DeleteFunc 指定一个 Delete 请求
func (p *Prefix) DeleteFunc(path string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(path, f, http.MethodDelete)
}

// PatchFunc 指定一个 Patch 请求
func (p *Prefix) PatchFunc(path string, f http.HandlerFunc) *Prefix {
	return p.HandleFunc(path, f, http.MethodPatch)
}
