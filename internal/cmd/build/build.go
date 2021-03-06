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
	v, err := versioninfo.New("./")
	if err != nil {
		return err
	}

	args := make([]string, 0, flagset.NArg()+1)
	args = append(args, "build")

	// flag 参数添加在最后，保证不会被其它设置顶替
	flag, err := v.LDFlags()
	if err != nil {
		return err
	}
	if flag != "" {
		args = append(args, "-ldflags", flag)
	}

	args = append(args, flagset.Args()...)

	cmd := exec.Command("go", args...)
	cmd.Stderr = output
	cmd.Stdout = output

	return cmd.Run()
}
