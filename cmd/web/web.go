// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

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

var subcommands = map[string]func() error{
	"version": version.Do,
	"help":    help.Do,
	"watch":   watch.Do,
	"create":  create.Do,
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

	fn()
}

func usage() {
	keys := make([]string, 0, len(subcommands))
	for k := range subcommands {
		keys = append(keys, k)
	}

	fmt.Printf(`web 命令是 github.com/issue9/web 框架提供的辅助工具。

目前支持以下子命令：%s
详情可以通过 web help 进行查看。
`, strings.Join(keys, ","))
}
