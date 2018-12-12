// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package watch

import (
	"flag"
	"fmt"
	"os"

	"github.com/caixw/gobuild"

	"github.com/issue9/web/internal/cmd/help"
)

func init() {
	help.Register("watch", usage)
}

// Do 执行子命令
func Do() error {
	var recursive, showIgnore bool
	var mainFiles, outputName, extString, appArgs string

	set := flag.NewFlagSet("", flag.ExitOnError)

	set.BoolVar(&recursive, "r", true, "是否查找子目录；")
	set.BoolVar(&showIgnore, "i", false, "是否显示被标记为 IGNORE 的日志内容；")
	set.StringVar(&outputName, "o", "", "指定输出名称，程序的工作目录随之改变；")
	set.StringVar(&appArgs, "x", "", "传递给编译程序的参数；")
	set.StringVar(&extString, "ext", "go", "指定监视的文件扩展，区分大小写。* 表示监视所有类型文件，空值代表不监视任何文件；")
	set.StringVar(&mainFiles, "main", "", "指定需要编译的文件；")
	set.Usage = usage
	set.Parse(os.Args[1:])

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
}
