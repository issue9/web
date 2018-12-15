// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package watch 提供热编译功能。
//
// 功能与 github.com/caixw/gobuild 相同。
package watch

import (
	"flag"
	"fmt"
	"os"

	"github.com/caixw/gobuild"

	"github.com/issue9/web/internal/cmd/help"
)

var (
	recursive, showIgnore                     bool
	mainFiles, outputName, extString, appArgs string

	flagset = flag.NewFlagSet("watch", flag.ExitOnError)
)

func init() {
	help.Register("watch", usage)

	flagset.BoolVar(&recursive, "r", true, "是否查找子目录；")
	flagset.BoolVar(&showIgnore, "i", false, "是否显示被标记为 IGNORE 的日志内容；")
	flagset.StringVar(&outputName, "o", "", "指定输出名称，程序的工作目录随之改变；")
	flagset.StringVar(&appArgs, "x", "", "传递给编译程序的参数；")
	flagset.StringVar(&extString, "ext", "go", "指定监视的文件扩展，区分大小写。* 表示监视所有类型文件，空值代表不监视任何文件；")
	flagset.StringVar(&mainFiles, "main", "", "指定需要编译的文件；")
	flagset.Usage = usage
}

// Do 执行子命令
func Do() error {
	flagset.Parse(os.Args[1:])

	logs := gobuild.NewConsoleLogs(showIgnore)

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dirs := append([]string{wd}, flag.Args()...)

	err = gobuild.Build(logs.Logs, mainFiles, outputName, extString, recursive, appArgs, dirs...)
	if err != nil {
		panic(err)
	}
	logs.Stop()
	return nil
}

func usage() {
	fmt.Println(`语法：web watch

热编译指定项目`)

	flagset.PrintDefaults()
}
