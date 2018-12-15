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

func usage() {
	fmt.Println(`构建一个新的 web 项目

语法：web create [mod]
mod 为一个可选参数，如果指定了，则会直接使用此值作为模块名，
若不指定，则会通过之后的交互要求用户指定。模块名中的最后一
路径名称，会作为目录名称创建于当前目录下。`)
}
