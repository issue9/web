// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package build 编译程，直接引用 go build
package build

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/issue9/cmdopt"
)

// Init 初始化函数
func Init(opt *cmdopt.CmdOpt) {
	opt.New("build", do, usage)
}

func do(output io.Writer) error {
	cmd := exec.Command("go", os.Args[1:]...)
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
