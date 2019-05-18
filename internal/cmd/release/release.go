// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package release 发布版本号管理
package release

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/issue9/cmdopt"
	"github.com/issue9/utils"
	"github.com/issue9/version"
)

// 指定版本化文件的路径
const path = "internal/version/version.go"

var flagset *flag.FlagSet

// Init 初始化函数
func Init(opt *cmdopt.CmdOpt) {
	flagset = opt.New("release", do, usage)
}

func do(output io.Writer) error {
	ver := flagset.Arg(0)

	if ver == "" {
		return errors.New("必须指定一个版本号")
	}

	if !version.SemVerValid(ver) {
		_, err := fmt.Fprintf(output, "无效的版本号格式：%s", ver)
		return err
	}

	// 输出到 internal/version/version.go
	dumpFile(ver)

	// 输出 git 标签
	cmd := exec.Command("git", "tag", "v"+ver)
	cmd.Stderr = output
	cmd.Stdout = output

	return cmd.Run()
}

func findAppRoot(curr string) (string, error) {
	path, err := filepath.Abs(curr)
	if err != nil {
		return "", err
	}

	for {
		if utils.FileExists(filepath.Join(path, "go.mod")) {
			return path, nil
		}

		p1 := filepath.Dir(path)
		if path == p1 {
			return "", errors.New("未找到根目录")
		}
		path = p1
	}
}

func dumpFile(ver string) error {
	root, err := findAppRoot("./")
	if err != nil {
		return err
	}

	p := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(p), os.ModePerm); err != nil {
		return err
	}

	return utils.DumpGoFile(p, fmt.Sprintf(versiongo, ver))
}

func usage(output io.Writer) error {
	_, err := fmt.Fprintf(output, `为当前程序发布一个新版本

该操作会在当前目录下添加 %s 文件，
并在其中写入版本信息。同时会通过 git tag 命令添加一条 tag 信息。
之后的 web build 会更新 %s 中的
buildDate 信息，但不会写入文件。

版本号的固定格式为 major.minjor.patch，比如 1.0.1，
git tag 标签中会自动加上 v 前缀，变成 v1.0.1。
`, path, path)

	return err
}
