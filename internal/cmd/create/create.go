// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package create

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/issue9/web/internal/cmd/help"
)

func init() {
	help.Register("create", usage)
}

// Do 执行子命令
func Do() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("模块名称")
	path, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	return create(path)
}

func usage() {
	fmt.Println(`语法：web create

构建一个新的 web 项目`)
}

func create(path string) error {
	name := filepath.Base(path)
	fmt.Println(name)
	// TODO
	return nil
}
