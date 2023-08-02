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

const (
	title = localeutil.StringPhrase("build go source")
	usage = localeutil.StringPhrase("build usage")
)

func Init(opt *cmdopt.CmdOpt, p *message.Printer) {
	opt.NewPlain("build", title.LocaleString(p), usage.LocaleString(p), build)
}

func build(w io.Writer, args []string) error {
	ver, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output()
	if err != nil {
		return err
	}

	fullCommit, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return err
	}

	commit, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return err
	}

	replaceVar(args, string(ver), string(fullCommit), string(commit))

	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// 替换变量
func replaceVar(args []string, ver, fullCommit, commit string) {
	r := strings.NewReplacer("{{version}}", ver, "{{full-commit}}", fullCommit, "{{commit}}", commit)
	for index, arg := range args {
		args[index] = r.Replace(arg)
	}
}
