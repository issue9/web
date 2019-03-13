// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package command 子命令管理
package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var commands = map[string]*command{}

// ErrNotFound 子命令未注册，通过 web hep [subcommnad]
// 时，返回此错误信息。
var ErrNotFound = errors.New("找不到该子命令")

// 每个子命令的结构
type command struct {
	// 命令的实际执行函数
	do func(io.Writer) error

	// 当前命令的帮助内容输出函数。
	usage func(io.Writer)
}

func init() {
	Register("help", helpDo, helpUsage)
}

// Exec 执行子命令
func Exec(output io.Writer) error {
	fn := usage
	if cmd, found := commands[os.Args[1]]; found {
		fn = cmd.do
	}

	return fn(output)
}

// Register 注册 usage 函数，注册的功能会在调用 web help xx 时调用。
func Register(name string, do func(io.Writer) error, usage func(io.Writer)) {
	if _, exists := commands[name]; exists {
		panic("存在同名的子命令:" + name)
	}

	commands[name] = &command{
		do:    do,
		usage: usage,
	}
}

func helpDo(output io.Writer) error {
	if len(os.Args) >= 3 {
		if cmd, found := commands[os.Args[2]]; found {
			cmd.usage(output)
			return nil
		}
	}

	return ErrNotFound
}

func helpUsage(output io.Writer) {
	fmt.Fprintln(output, `显示名子命令的相关介绍

用法：web help [subcommand]`)
}

func usage(output io.Writer) error {
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
