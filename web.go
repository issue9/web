// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"runtime"
	"strings"

	"github.com/issue9/version"
	"github.com/issue9/web/app"
	"github.com/issue9/web/context"
	"github.com/issue9/web/module"
	"github.com/issue9/web/result"
)

const (
	// Version 当前框架的版本
	Version = "0.17.0+20181118"

	// MinimumGoVersion 需求的最低 Go 版本
	// 修改此值，记得同时修改 .travis.yml 文件中的版本依赖。
	MinimumGoVersion = "1.11"
)

type (
	// Context 等同于 context.Context，方便调用者使用
	Context = context.Context

	// Result 等同于 result.Result，方便调用者使用
	Result = result.Result

	// Module 等同于 module.Module，方便调用者使用
	Module = module.Module

	// Options Init 的参数类型
	Options = app.Options
)

// NewResult 生成一个 *result.Result 对象
func NewResult(code int) *Result {
	return &result.Result{Code: code}
}

// 作最低版本检测
func init() {
	checkVersion(runtime.Version())
}

func checkVersion(goversion string) {
	goversion = strings.TrimPrefix(goversion, "go")

	// tip 版本，不作检测
	if strings.HasPrefix(goversion, "devel ") {
		return
	}

	v, err := version.SemVerCompare(goversion, MinimumGoVersion)
	if err != nil {
		panic(err)
	}

	if v < 0 {
		panic("低于最小版本需求")
	}
}
