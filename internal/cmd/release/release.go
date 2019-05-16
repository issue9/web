// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package release 发布版本号管理
package release

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/issue9/utils"
	"github.com/issue9/version"
)

// 指定版本化文件的路 径
const path = "internal/version/version.go"

// Do 执行子命令
func Do(output io.Writer) error {
	ver := os.Args[2]

	if !version.SemVerValid(ver) {
		_, err := fmt.Fprintln(output, "无效的版本号格式！")
		return err
	}

	// 输出到 internal/version/version.go
	dumpFile(ver)

	// 输出 git 标签
	cmd := exec.Command("git", "tag", "v"+ver)
	cmd.Stderr = output
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
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

// Usage 当前子命令的用法
func Usage(output io.Writer) {
	fmt.Fprintln(output, `为当前程序发布一个新版本

将会执行以下操作：
1 添加新的 git tag；
2 更新本地代码的版本号。`)
}
