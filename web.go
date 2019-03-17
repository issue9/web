// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"github.com/issue9/web/context"
	"github.com/issue9/web/module"
)

const (
	// Version 当前框架的版本
	Version = "0.25.3"

	// MinimumGoVersion 需求的最低 Go 版本
	// 修改此值，记得同时修改 .travis.yml 文件中的版本依赖以及 README.md 中的相关信息。
	MinimumGoVersion = "1.11"

	// CoreModuleName 框架自带的模块名称
	CoreModuleName = module.CoreModuleName
)

type (
	// Context 等同于 context.Context，方便调用者使用
	Context = context.Context

	// Result 等同于 context.Resut，方便调用者使用
	Result = context.Result

	// Module 等同于 module.Module，方便调用者使用
	Module = module.Module

	// Service 等同于 module.Service，方便调用者使用
	Service = module.Service

	// ServiceFunc 等同于 module.ServiceFunc，方便调用者使用
	ServiceFunc = module.ServiceFunc
)
