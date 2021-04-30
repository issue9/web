// SPDX-License-Identifier: MIT

// Package build 提供编译相关的功能
package build

import (
	"flag"
	"io"
	"os/exec"

	"github.com/issue9/cmdopt"

	"github.com/issue9/web/internal/versioninfo"
)

const usage = `编译当前程序，功能与 go build 完全相同！

如果你使用 web release 发布了版本号，则当前操作还会在每一次编译时指定个编译日期，
固定格式为 YYYYMMDD。
`

var flagset *flag.FlagSet

// Init 初始化函数
func Init(opt *cmdopt.CmdOpt) {
	flagset = opt.New("build", usage, do)
}

func do(output io.Writer) error {
	dir, err := versioninfo.Root("./")
	if err != nil {
		return err
	}

	if err := dir.DumpVersionFile("./"); err != nil {
		return err
	}

	// build

	args := make([]string, 0, flagset.NArg()+1)
	args = append(args, "build")

	args = append(args, flagset.Args()...)

	cmd := exec.Command("go", args...)
	cmd.Stderr = output
	cmd.Stdout = output

	return cmd.Run()
}
