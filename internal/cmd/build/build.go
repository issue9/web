// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package build 编译程，直接引用 go build
package build

import (
	"flag"
	"fmt"
	"io"
	"os/exec"

	"github.com/issue9/cmdopt"

	"github.com/issue9/web/internal/versioninfo"
)

var flagset *flag.FlagSet

// Init 初始化函数
func Init(opt *cmdopt.CmdOpt) {
	flagset = opt.New("build", do, usage)
}

func do(output io.Writer) error {
	args := make([]string, 0, len(flagset.Args())+1)
	args = append(args, flagset.Args()...)

	v, err := versioninfo.New("./")
	if err != nil {
		return err
	}

	// flag 参数添加在最后，保证不会被其它设置顶替
	flag, err := v.LDFlags()
	if err != nil {
		return err
	}
	if flag != "" {
		args = append(args, "-ldflags", flag)
	}

	cmd := exec.Command("go", args...)
	cmd.Stderr = output
	cmd.Stdout = output

	return cmd.Run()
}

func usage(output io.Writer) error {
	_, err := fmt.Fprintln(output, `编译当前程序，功能与 go build 完全相同！

如果你使用 web release 发布了版本号，则当前操作还会在每一次编译时指定个编译日期，
固定格式为 YYYYMMDD。`)
	return err
}
