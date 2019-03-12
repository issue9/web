// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package command 子命令管理
package command

import (
	"fmt"
	"os"
	"strings"
)

var commands = map[string]*command{}

// 每个子命令的结构
type command struct {
	// 命令的实际执行函数
	do func(*os.File) error

	// 当前命令的帮助内容输出函数。
	usage func(*os.File)
}

func init() {
	Register("help", helpDo, helpUsage)
}

// Exec 执行子命令
func Exec(output *os.File) {
	fn := usage
	if cmd, found := commands[os.Args[1]]; found {
		fn = cmd.do
	}

	fn(output)
}

func helpDo(output *os.File) error {
	fn := helpNotExistsUsage
	if len(os.Args) >= 3 {
		if cmd, found := commands[os.Args[2]]; found {
			fn = cmd.usage
		}
	}

	fn(output)
	return nil
}

// Register 注册 usage 函数，注册的功能会在调用 web help xx 时调用。
func Register(name string, do func(*os.File) error, usage func(*os.File)) {
	if _, exists := commands[name]; exists {
		panic("存在同名的子命令:" + name)
	}

	commands[name] = &command{
		do:    do,
		usage: usage,
	}
}

func helpUsage(output *os.File) {
	fmt.Fprintln(output, `显示名子命令的相关介绍

用法：web help [subcommand]`)
}

func helpNotExistsUsage(output *os.File) {
	// TODO
}

func usage(output *os.File) error {
	keys := make([]string, 0, len(commands))
	for k := range commands {
		keys = append(keys, k)
	}

	_, err := fmt.Fprintf(output, `web 命令是 github.com/issue9/web 框架提供的辅助工具。

目前支持以下子命令：%s
详情可以通过 web help [subcommand] 进行查看。
`, strings.Join(keys, ","))

	return err
}
