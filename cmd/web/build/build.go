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
	"github.com/issue9/term/v3/colors"
	"github.com/issue9/web"
	"golang.org/x/text/message"

	"github.com/issue9/web/cmd/web/git"
)

const (
	title = localeutil.StringPhrase("build go source")
	usage = localeutil.StringPhrase("build usage")
)

func Init(opt *cmdopt.CmdOpt, p *message.Printer) {
	opt.NewPlain("build", title.LocaleString(p), usage.LocaleString(p), func(w io.Writer, args []string) error {
		ver, err := git.Version()
		if err != nil {
			colors.Println(colors.Normal, colors.Yellow, colors.Default, err.(web.LocaleStringer).LocaleString(p))
		}

		commit, err := git.Commit(false)
		if err != nil {
			colors.Println(colors.Normal, colors.Yellow, colors.Default, err.(web.LocaleStringer).LocaleString(p))
		}

		commitFull, err := git.Commit(true)
		if err != nil {
			colors.Println(colors.Normal, colors.Yellow, colors.Default, err.(web.LocaleStringer).LocaleString(p))
		}

		replaceVar(args, ver, commitFull, commit)

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
