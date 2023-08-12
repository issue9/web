// SPDX-License-Identifier: MIT

// Package git Git 命令行操作
package git

import (
	"bytes"
	"os/exec"

	"github.com/issue9/localeutil"
	"github.com/issue9/term/v3/colors"
	"github.com/issue9/web"
)

// Run 运行 git 命令
func Run(p *localeutil.Printer, presetValue string, args ...string) string {
	cmd := exec.Command("git", args...)
	buf := &bytes.Buffer{}
	cmd.Stderr = buf

	output, err := cmd.Output()
	if err != nil {
		p := web.Phrase("%s when exec %s, use the preset value %s", buf.String(), cmd.String(), presetValue).LocaleString(p)
		colors.Println(colors.Normal, colors.Yellow, colors.Default, p)
		return presetValue
	}
	return string(bytes.TrimSpace(output))
}

// Version 返回当前 git 仓库的最新标签
func Version(p *localeutil.Printer) string {
	return Run(p, "dev", "describe", "--tags", "--abbrev=0")
}

// Commit 最后一次的 git 提交 hash
func Commit(p *localeutil.Printer, full bool) string {
	if full {
		return Run(p, "", "rev-parse", "HEAD")
	}
	return Run(p, "", "rev-parse", "--short", "HEAD")
}
