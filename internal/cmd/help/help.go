// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package help 管理 web help xx 的显示信息
package help

import (
	"fmt"
	"os"
)

var usages = map[string]func(){}

func init() {
	Register("help", usage)
}

// Do 执行子命令
func Do() error {
	fn := usage
	if len(os.Args) >= 3 {
		if f, found := usages[os.Args[2]]; found {
			fn = f
		}
	}

	fn()
	return nil
}

// Register 注册 usage 函数，注册的功能会在调用 web help xx 时调用。
func Register(name string, fn func()) {
	if _, exists := usages[name]; exists {
		panic(fmt.Sprintln("存在同名的子命令:", name))
	}

	usages[name] = fn
}

func usage() {
	fmt.Println(`显示名子命令的相关介绍

用法：web help subcommand`)
}
