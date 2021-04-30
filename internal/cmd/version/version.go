// SPDX-License-Identifier: MIT

// Package version 显示版本号信息
package version

import (
	"fmt"
	"io"
	"runtime"

	"github.com/issue9/cmdopt"

	"github.com/issue9/web/internal/version"
)

const usage = "显示当前程序的版本号\n"

// Init 初始化函数
func Init(opt *cmdopt.CmdOpt) {
	opt.New("version", usage, printLocalVersion)
}

func printLocalVersion(output io.Writer) error {
	v := version.Version
	d := version.Date
	c := version.Commits
	h := version.Hash
	_, err := fmt.Fprintf(output, "web: %s+%s.%d.%s\nbuild with %s\n", v, d, c, h, runtime.Version())
	return err
}
