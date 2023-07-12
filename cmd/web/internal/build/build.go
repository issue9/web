// SPDX-License-Identifier: MIT

// Package build 提供 build 子命令
package build

import (
	"flag"
	"io"
	"os/exec"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"golang.org/x/text/message"
)

var (
	title = localeutil.Phrase("build go source with version from git tag")
	usage = localeutil.Phrase("build usage")
)

func Init(opt *cmdopt.CmdOpt, p *message.Printer) {
	opt.New("build", title.LocaleString(p), usage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		return func(w io.Writer) error {
			//
		}
	})
}

func getLatestTag(src string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
