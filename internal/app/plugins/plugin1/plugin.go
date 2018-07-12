// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

//go:generate go build -o=../plugin1.so -buildmode=plugin ./plugin.go

// 测试用的 plugin 模块，可以直接运行 go generate 生成 .so 文件
package main

import (
	"fmt"

	"github.com/issue9/web/module"
)

// Init 初始化模块
func Init(m *module.Module) {
	m.Name = "plugin1"
	m.Description = "plugin1 desc"
	m.Deps = []string{"plugin2"} // 依赖插件 2

	m.AddInit(init1)
	m.AddInit(init2)

	m.NewVersion("install").Task("title", install1)
	m.NewVersion("v1.0").Task("title", install2)
}

func init1() error {
	fmt.Println("plugin1 init1")
	return nil
}

func init2() error {
	fmt.Println("plugin1 init2")
	return nil
}

func install1() error {
	fmt.Println("plugin1 install1")
	return nil
}

func install2() error {
	fmt.Println("plugin1 install2")
	return nil
}

func main() {}
