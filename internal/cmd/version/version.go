// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package version

import (
	"fmt"
	"runtime"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/cmd/help"
)

func init() {
	help.Register("version", usage)
}

// Do 执行子命令
func Do() error {
	_, err := fmt.Printf("web:%s build with %s\n", web.Version, runtime.Version())
	return err
}

func usage() {
	fmt.Println(`语法：web version

显示当前程序的版本号`)
}
