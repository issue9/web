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

const usage = `编译当前程序

在未指定 version 参数的情况下，其功能与 go build 完全相同，
如果指定了 version 参数，则会输出 internal/version/info.go
文件，并在该文件中包含相关的版本信息。
`

var (
	flagset *flag.FlagSet
	v       bool
)

// Init 初始化函数
func Init(opt *cmdopt.CmdOpt) {
	flagset = opt.New("build", usage, do)
	flagset.BoolVar(&v, "version", false, "按 web 的格式输出相关版本信息")
}

func do(output io.Writer) error {
	dir, err := versioninfo.Root("./")
	if err != nil {
		return err
	}

	if v {
		if err := dir.DumpFile(); err != nil {
			return err
		}
	}

	args := make([]string, 0, flagset.NArg()+1)
	args = append(args, "build")

	args = append(args, flagset.Args()...)

	cmd := exec.Command("go", args...)
	cmd.Stderr = output
	cmd.Stdout = output

	return cmd.Run()
}
