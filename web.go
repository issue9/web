// SPDX-License-Identifier: MIT

package web

import (
	"net/http"

	"github.com/issue9/config"
	"github.com/issue9/scheduled"

	"github.com/issue9/web/app"
	"github.com/issue9/web/context"
	"github.com/issue9/web/internal/version"
	"github.com/issue9/web/module"
)

// Version 当前框架的版本
const Version = version.Version

type (
	// Context 一般在 http.ServeHTTP 中根据 http.ResponseWriter
	// 和 http.Request 初始化获得。
	// 可以在整个函数生命周期中操作相关功能。
	//
	// 等同于 context.Context，方便调用者使用
	Context = context.Context

	// Result 等同于 app.Result，方便调用者使用
	Result = app.Result

	// ServiceFunc 服务的报告函数签名。
	//
	// 等同于 app.ServiceFunc，方便调用者使用
	ServiceFunc = app.ServiceFunc

	// SchedulerFunc 计划任务的执行函数签名。
	//
	// 等同于 scheduled.JobFunc，方便调用者使用
	SchedulerFunc = scheduled.JobFunc

	// Module 等同于 module.Module，方便调用者使用
	Module = module.Module

	// Tag 等同于 module.Tag，方便调用者使用
	Tag = module.Tag

	// ConfigManager 配置管理
	//
	// 管理不同类型的文件加载，等同于 github.com/issue9/config.Manager
	ConfigManager = config.Manager
)

// NewConfigManager 声明 ConfigManager 实例
func NewConfigManager(dir string) (*ConfigManager, error) {
	return config.NewManager(dir)
}

// NewContext 生成 *Context 对象，若是出错则 panic
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return context.New(w, r, App())
}
