// SPDX-License-Identifier: MIT

// Package build 提供 build 子命令
package build

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/term/v3/colors"
	"github.com/issue9/web"
	"golang.org/x/text/message"
)

const (
	title = localeutil.StringPhrase("build go source")
	usage = localeutil.StringPhrase("build usage")
)

func Init(opt *cmdopt.CmdOpt, p *message.Printer) {
	opt.NewPlain("build", title.LocaleString(p), usage.LocaleString(p), func(w io.Writer, args []string) error {
		ver := runGit(p, "dev", "describe", "--tags", "--abbrev=0")
		fullCommit := runGit(p, "", "rev-parse", "HEAD")
		commit := runGit(p, "", "rev-parse", "--short", "HEAD")

		replaceVar(args, ver, fullCommit, commit)

		cmd := exec.Command("go", append([]string{"build"}, args...)...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		return cmd.Run()
	})
}

// 替换变量
func replaceVar(args []string, ver, fullCommit, commit string) {
	r := strings.NewReplacer("{{version}}", ver, "{{full-commit}}", fullCommit, "{{commit}}", commit)
	for index, arg := range args {
		args[index] = r.Replace(arg)
	}
}

func runGit(p *localeutil.Printer, presetValue string, args ...string) string {
	cmd := exec.Command("git", args...)
	buf := &bytes.Buffer{}
	cmd.Stderr = buf

	output, err := cmd.Output()
	if err != nil {
		p := web.Phrase("%s when exec %s, use the preset value %s", buf.String(), cmd.String(), presetValue).LocaleString(p)
		colors.Println(colors.Normal, colors.Yellow, colors.Default, p)
		return presetValue
	}
	return string(output)
}
