// SPDX-License-Identifier: MIT

// 简单的辅助功能命令行工具
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/issue9/cmdopt"

	"github.com/issue9/web/internal/cmd/build"
	"github.com/issue9/web/internal/cmd/create"
	"github.com/issue9/web/internal/cmd/release"
	"github.com/issue9/web/internal/cmd/version"
	"github.com/issue9/web/internal/cmd/watch"
)

var opt *cmdopt.CmdOpt

func main() {
	opt = cmdopt.New(os.Stdout, flag.ExitOnError, header, footer, "选项：", "子命令：", func(name string) string {
		return fmt.Sprintf("未找到子命令 %s", name)
	})

	opt.Help("help", "显示当前内容\n")
	version.Init(opt)
	build.Init(opt)
	create.Init(opt)
	watch.Init(opt)
	release.Init(opt)

	if err := opt.Exec(os.Args[1:]); err != nil {
		panic(err)
	}
}

const header = `web 命令是 github.com/issue9/web 框架提供的辅助工具。
`

const footer = `当前项目源码以 MIT 许可发布于 https://github.com/issue9/web
`
