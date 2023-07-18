// SPDX-License-Identifier: MIT

// Package build 提供 build 子命令
package build

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"golang.org/x/text/message"
)

var (
	title = localeutil.Phrase("build go source with version from git tag")
	usage = localeutil.Phrase("build usage")
)

func Init(opt *cmdopt.CmdOpt, p *message.Printer) {
	opt.NewPlain("build", title.LocaleString(p), usage.LocaleString(p), build)
}

func build(w io.Writer, args []string) error {
	ver, err := getLatestTag(args[len(args)-1])
	if err != nil {
		return err
	}

	replaceVar(args, ver)

	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// 替换变量
//
// 目前支持以下变量：
//
//   - {{version}}
func replaceVar(args []string, ver string) {
	for index, arg := range args {
		arg = strings.ReplaceAll(arg, "{{version}}", ver)
		args[index] = arg
	}
}

func getLatestTag(src string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
