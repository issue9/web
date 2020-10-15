// SPDX-License-Identifier: MIT

// Package cmd 命令行相关操作
package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/issue9/cmdopt"

	"github.com/issue9/web/internal/cmd/build"
	"github.com/issue9/web/internal/cmd/release"
	"github.com/issue9/web/internal/cmd/version"
	"github.com/issue9/web/internal/cmd/watch"
)

const (
	header = `web 命令是 github.com/issue9/web 框架提供的辅助工具
`

	footer = `当前项目源码以 MIT 许可发布于 https://github.com/issue9/web
`
)

// Exec 执行命令行操作
func Exec() error {
	opt := &cmdopt.CmdOpt{
		Output:        os.Stdout,
		ErrorHandling: flag.ExitOnError,
		Header:        header,
		Footer:        footer,
		OptionsTitle:  "选项：",
		CommandsTitle: "子命令：",
		NotFound: func(name string) string {
			return fmt.Sprintf("未找到子命令 %s", name)
		},
	}

	opt.Help("help", "显示当前内容\n")
	version.Init(opt)
	build.Init(opt)
	watch.Init(opt)
	release.Init(opt)

	return opt.Exec(os.Args[1:])
}
