// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package create

import (
	"fmt"

	"github.com/issue9/web/internal/cmd/help"
)

func init() {
	help.Register("create", usage)
}

// Do 执行子命令
func Do() error {
	// TODO
	return nil
}

func usage() {
	fmt.Println(`语法：web create

构建一个新的 web 项目`)
}
