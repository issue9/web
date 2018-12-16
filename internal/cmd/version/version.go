// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package version 显示版本号信息
package version

import (
	"fmt"
	"os"
	"runtime"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/cmd/help"
)

func init() {
	help.Register("version", usage)
}

// Do 执行子命令
func Do(output *os.File) error {
	_, err := fmt.Fprintf(output, "web:%s build with %s\n", web.Version, runtime.Version())
	return err
}

func usage(output *os.File) {
	fmt.Fprintln(output, `显示当前程序的版本号

语法：web version`)
}
