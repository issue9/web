// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 简单的辅助功能命令行工具。
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/issue9/web/internal/cmd/create"
	"github.com/issue9/web/internal/cmd/help"
	"github.com/issue9/web/internal/cmd/version"
	"github.com/issue9/web/internal/cmd/watch"
)

// 帮助信息的输出通道
var output = os.Stdout

var subcommands = map[string]func(*os.File) error{
	"version": version.Do,
	"watch":   watch.Do,
	"create":  create.Do,
	"help":    help.Do,
}

func main() {
	if len(os.Args) == 1 {
		usage()
		return
	}

	fn, found := subcommands[os.Args[1]]
	if !found {
		usage()
		return
	}

	fn(output)
}

func usage() {
	keys := make([]string, 0, len(subcommands))
	for k := range subcommands {
		keys = append(keys, k)
	}

	fmt.Fprintf(output, `web 命令是 github.com/issue9/web 框架提供的辅助工具。

目前支持以下子命令：%s
详情可以通过 web help 进行查看。
`, strings.Join(keys, ","))
}
