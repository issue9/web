// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 测试用的 plugin 模块，可以直接运行 go generate 生成 .so 文件
package main

import "github.com/issue9/web/module"

// Init 初始化模块
func Init(m *module.Module) {
	//m.Name = "xx"
	// NOTE(caixw): 这是个错误的 plugin，缺少必要的名称设置
}

func main() {}
