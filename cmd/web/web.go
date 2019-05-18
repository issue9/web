// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 简单的辅助功能命令行工具。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/issue9/cmdopt"

	"github.com/issue9/web/internal/cmd/build"
	"github.com/issue9/web/internal/cmd/create"
	"github.com/issue9/web/internal/cmd/release"
	"github.com/issue9/web/internal/cmd/version"
	"github.com/issue9/web/internal/cmd/watch"
)

var opt *cmdopt.CmdOpt

func main() {
	opt = cmdopt.New(os.Stdout, flag.ExitOnError, usage)

	opt.Help("help")
	version.Init(opt)
	build.Init(opt)
	create.Init(opt)
	watch.Init(opt)
	release.Init(opt)

	if err := opt.Exec(os.Args[1:]); err != nil {
		panic(err)
	}
}

func usage(output io.Writer) error {
	_, err := fmt.Fprintf(output, `web 命令是 github.com/issue9/web 框架提供的辅助工具。

目前支持以下子命令：%s
详情可以通过 web help [subcommand] 进行查看。
`, strings.Join(opt.Commands(), ","))

	return err
}
