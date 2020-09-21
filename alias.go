// SPDX-License-Identifier: MIT

package web

import (
	"github.com/issue9/scheduled"

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

	// Result 等同于 context.Result，方便调用者使用
	Result = context.Result

	// ServiceFunc 服务的报告函数签名
	//
	// 等同于 server.ServiceFunc，方便调用者使用
	ServiceFunc = module.ServiceFunc

	// SchedulerFunc 计划任务的执行函数签名
	//
	// 等同于 scheduled.JobFunc，方便调用者使用
	SchedulerFunc = scheduled.JobFunc

	// Module 等同于 server.Module，方便调用者使用
	Module = module.Module

	// Tag 等同于 server.Tag，方便调用者使用
	Tag = module.Tag
)
