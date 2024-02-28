// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package git Git 命令行操作
package git

import (
	"bytes"
	"os/exec"

	"github.com/issue9/web"
)

// Run 运行 git 命令
//
// 返回的错误类型始终可以转换成 [web.LocaleStringer] 类型。
func Run(presetValue string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	buf := &bytes.Buffer{}
	cmd.Stderr = buf

	output, err := cmd.Output()
	if err != nil {
		return presetValue, web.NewLocaleError("%s when exec %s, use the preset value %s", buf.String(), cmd.String(), presetValue)
	}
	return string(bytes.TrimSpace(output)), nil
}

// Version 返回当前 git 仓库的最新标签
func Version() (string, error) {
	return Run("dev", "describe", "--tags", "--abbrev=0")
}

// Commit 最后一次的 git 提交 hash
func Commit(full bool) (string, error) {
	if full {
		return Run("", "rev-parse", "HEAD")
	}
	return Run("", "rev-parse", "--short", "HEAD")
}
