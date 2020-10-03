// SPDX-License-Identifier: MIT

package web

import (
	"github.com/issue9/middleware/v2"
	"github.com/issue9/web/context"
	"github.com/issue9/web/internal/version"
	"github.com/issue9/web/module"
)

// Version 当前框架的版本
const Version = version.Version

type (
	// Context 定义了在单个 HTTP 请求期间的上下文环境
	//
	// 是对 http.ResponseWriter 和 http.Request 的简单包装。
	Context = context.Context

	// Middleware 中间件的类型定义
	Middleware = middleware.Middleware

	// Result 定义了返回给用户的错误信息
	Result = context.Result

	// SchedulerFunc 计划任务的执行函数签名
	SchedulerFunc = module.JobFunc

	// Module 定义了模块的相关信息
	Module = module.Module

	// Service 常驻后台运行的服务描述
	Service = module.Service

	// ServiceFunc 服务的执行函数签名
	ServiceFunc = module.ServiceFunc

	// MODServer 模块的管理服务
	MODServer = module.Server
)
