// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"github.com/issue9/scheduled"
	"github.com/issue9/web/app"
	"github.com/issue9/web/context"
)

const (
	// Version 当前框架的版本
	Version = "0.26.0"

	// MinimumGoVersion 需求的最低 Go 版本
	// 修改此值，记得同时修改 .travis.yml 文件中的版本依赖以及 README.md 中的相关信息。
	MinimumGoVersion = "1.11"

	// CoreModuleName 框架自带的模块名称
	CoreModuleName = app.CoreModuleName
)

type (
	// Context 等同于 context.Context，方便调用者使用
	Context = context.Context

	// Result 等同于 app.Resut，方便调用者使用
	Result = app.Result

	// Module 等同于 app.Module，方便调用者使用
	Module = app.Module

	// Service 等同于 app.Service，方便调用者使用
	Service = app.Service

	// ServiceFunc 等同于 app.ServiceFunc，方便调用者使用
	ServiceFunc = app.ServiceFunc

	// Scheduler 等同于 scheduled.Job，方便调用者使用
	Scheduler = scheduled.Job
)
