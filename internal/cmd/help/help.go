// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package help

import (
	"errors"
	"fmt"
	"os"
)

var errNotExists = errors.New("不存在子命令")

var usages = map[string]func(){
	"help": usage,
}

// Do 执行子命令
func Do() error {
	if len(os.Args) == 1 {
		return errNotExists
	}

	usage, found := usages[os.Args[1]]
	if !found {
		return errNotExists
	}

	usage()

	return nil
}

// Register 注册 usage 函数
func Register(name string, usage func()) {
	if _, exists := usages[name]; exists {
		panic(fmt.Sprintln("存在同名的子命令:", name))
	}

	usages[name] = usage
}


func usage() {
	fmt.Println(`用法：web help subcommand

显示名为 subcommand 的子命令的用法。`)
}

