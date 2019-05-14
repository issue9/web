// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package release

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Do 执行子命令
func Do(output io.Writer) error {
	tag := os.Args[2]

	// TODO 更新版本号

	cmd := exec.Command("git", "tag", tag)
	cmd.Stderr = output
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// Usage 当前子命令的用法
func Usage(output io.Writer) {
	fmt.Fprintln(output, `为当前程序发布一个新版本

将会执行以下操作：
1 添加新的 git tag；
2 更新本地代码的版本号。`)
}
