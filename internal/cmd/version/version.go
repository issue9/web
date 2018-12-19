// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package version 显示版本号信息
package version

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/cmd/help"
)

// 用于获取版本信息的 git 仓库地址
const repoURL = "https://github.com/issue9/web"

var (
	localVersion = web.Version
	buildDate    string
)

var (
	check   bool
	flagset *flag.FlagSet
)

func init() {
	help.Register("version", usage)

	if buildDate != "" {
		localVersion += ("+" + buildDate)
	}

	flagset = flag.NewFlagSet("version", flag.ExitOnError)
	flagset.BoolVar(&check, "c", false, "是否检测线上的最新版本")
}

// Do 执行子命令
func Do(output *os.File) error {
	if err := flagset.Parse(os.Args[2:]); err != nil {
		return err
	}

	if check {
		return checkRemoteVersion(output)
	}

	_, err := fmt.Fprintf(output, "web:%s build with %s\n", localVersion, runtime.Version())
	return err
}

// 检测框架的最新版本号
//
// 获取线上的标签列表，拿到其中的最大值。
func checkRemoteVersion(output *os.File) error {
	cmd := exec.Command("git", "ls-remote", "--tags", repoURL)
	buf := new(bytes.Buffer)
	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		return err
	}

	ver, err := getMaxVersion(buf)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(output, "local:%s build with %s\n", localVersion, runtime.Version())
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(output, "latest:%s\n", ver)
	return err
}

func getMaxVersion(buf *bytes.Buffer) (string, error) {
	s := bufio.NewScanner(buf)
	var max string

	for s.Scan() {
		text := s.Text()
		index := strings.LastIndex(text, "/v")
		if index < 0 {
			continue
		}
		ver := text[index+2:]

		if ver > max {
			max = ver
		}
	}

	return max, nil
}

func usage(output *os.File) {
	fmt.Fprintln(output, `显示当前程序的版本号

语法：web version [options]
options`)
	flagset.SetOutput(output)
	flagset.PrintDefaults()
}
