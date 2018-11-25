// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"runtime"
	"strings"

	"github.com/issue9/version"
)

const (
	// Version 当前框架的版本
	Version = "0.17.0+20181118"

	// MinimumGoVersion 需求的最低 Go 版本
	// 修改此值，记得同时修改 .travis.yml 文件中的版本依赖。
	MinimumGoVersion = "1.11"
)

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
