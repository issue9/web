// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"github.com/issue9/scheduled"

	"github.com/issue9/web/app"
	"github.com/issue9/web/context"
	"github.com/issue9/web/module"
)

// Version 当前框架的版本
const Version = "0.26.0"

type (
	// Context 等同于 context.Context，方便调用者使用
	Context = context.Context

	// Result 等同于 app.Resut，方便调用者使用
	Result = app.Result

	// Service 等同于 app.Service，方便调用者使用
	Service = app.Service

	// ServiceFunc 等同于 app.ServiceFunc，方便调用者使用
	ServiceFunc = app.ServiceFunc

	// Scheduler 等同于 scheduled.Job，方便调用者使用
	Scheduler = scheduled.Job

	// Module 等同于 module.Module，方便调用者使用
	Module = module.Module
)
