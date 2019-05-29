// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package release 发布版本号管理
package release

import (
	"flag"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/issue9/cmdopt"
	"github.com/issue9/version"

	"github.com/issue9/web/internal/versioninfo"
)

var flagset *flag.FlagSet

// Init 初始化函数
func Init(opt *cmdopt.CmdOpt) {
	flagset = opt.New("release", do, usage)
}

func do(output io.Writer) error {
	ver := flagset.Arg(1)

	// 没有多余的参数，则会显示当前已有的版本号列表
	if ver == "" {
		return outputTags(output)
	}

	if ver[0] == 'v' || ver[0] == 'V' {
		ver = ver[1:]
	}

	if !version.SemVerValid(ver) {
		_, err := fmt.Fprintf(output, "无效的版本号格式：%s", ver)
		return err
	}

	v, err := versioninfo.New("./")
	if err != nil {
		return err
	}

	if err := v.DumpFile(ver); err != nil {
		return err
	}

	// 没有提交消息，则不提交内容到 VCS
	if len(flagset.Args()) <= 2 {
		return nil
	}

	var message string
	message = strings.Join(flagset.Args()[:2], " ")

	// 添加到 git 缓存中
	cmd := exec.Command("git", "add", filepath.Join(v.Path(versioninfo.Path)))
	cmd.Stderr = output
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Stderr = output
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		return err
	}

	// 输出 git 标签
	cmd = exec.Command("git", "tag", "v"+ver)
	cmd.Stderr = output
	cmd.Stdout = output

	return cmd.Run()
}

func outputTags(output io.Writer) error {
	cmd := exec.Command("git", "tag")
	cmd.Stdout = output

	return cmd.Run()
}

func usage(output io.Writer) error {
	_, err := fmt.Fprintf(output, `为当前程序发布一个新版本

该操作会在项目的根目录下添加 %s 文件，
并在其中写入版本信息。之后通过 web build 编译，
会更新 %s 中的 buildDate 信息，但不会写入文件。
同时根据参数决定是否用 git tag 命令添加一条 tag 信息。


版本号的固定格式为 major.minjor.patch，比如 1.0.1，当然 v1.0.1 也会被正确处理。
git tag 标签中会自动加上 v 前缀，变成 v1.0.1。

一般用法：
web release 0.1.1 [commit message]
`, versioninfo.Path, versioninfo.Path)

	return err
}
