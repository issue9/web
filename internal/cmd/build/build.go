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
)

// Do 执行子命令
func Do(output io.Writer) error {
	cmd := exec.Command("go", os.Args[1:]...)
	cmd.Stderr = output
	cmd.Stdout = output

	return cmd.Run()
}

// Usage 当前子命令的用法
func Usage(output io.Writer) {
	fmt.Fprintln(output, `编译当前程序，功能与 go build 完全相同！`)
}
