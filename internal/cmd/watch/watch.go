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
	"io"
	"os"

	"github.com/caixw/gobuild"
	"github.com/issue9/cmdopt"

	"github.com/issue9/web/internal/versioninfo"
)

var (
	recursive, showIgnore                     bool
	mainFiles, outputName, extString, appArgs string

	flagset *flag.FlagSet
)

// Init 初始化函数
func Init(opt *cmdopt.CmdOpt) {
	flagset = opt.New("watch", do, usage)
	flagset.BoolVar(&recursive, "r", true, "是否查找子目录；")
	flagset.BoolVar(&showIgnore, "i", false, "是否显示被标记为 IGNORE 的日志内容；")
	flagset.StringVar(&outputName, "o", "", "指定输出名称，程序的工作目录随之改变；")
	flagset.StringVar(&appArgs, "x", "", "传递给编译程序的参数；")
	flagset.StringVar(&extString, "ext", "go", "指定监视的文件扩展，区分大小写。* 表示监视所有类型文件，空值代表不监视任何文件；")
	flagset.StringVar(&mainFiles, "main", "", "指定需要编译的文件；")
}

func do(output io.Writer) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	dirs := append([]string{wd}, flag.Args()...)

	v, err := versioninfo.New("./")
	if err != nil {
		return err
	}
	ld, err := v.LDFlags()
	if err != nil {
		return err
	}
	flags := map[string]string{"ld": ld}

	logs := gobuild.NewConsoleLogs(showIgnore)
	defer logs.Stop()
	return gobuild.Build(logs.Logs, mainFiles, outputName, flags, extString, recursive, appArgs, dirs...)
}

func usage(output io.Writer) error {
	_, err := fmt.Fprintln(output, `热编译当前目录下的项目

命令行语法：
 web watch [options] [dependents]

 options:`)

	if err != nil {
		return err
	}

	flagset.PrintDefaults()

	_, err = fmt.Fprintln(output, `
 dependents:
  指定其它依赖的目录，只能出现在命令的尾部。


常见用法:

 web watch 
   监视当前目录，若有变动，则重新编译当前目录下的 *.go 文件；

 web watch -main=main.go
   监视当前目录，若有变动，则重新编译当前目录下的 main.go 文件；

 web watch -main="main.go" dir1 dir2
   监视当前目录及 dir1 和 dir2，若有变动，则重新编译当前目录下的 main.go 文件；


NOTE: 不会监视隐藏文件和隐藏目录下的文件。`)

	return err
}
