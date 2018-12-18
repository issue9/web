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
	"sort"
	"strings"

	"github.com/issue9/version"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/cmd/help"
)

var (
	localVersion = web.Version
	buildDate    string
)

func init() {
	help.Register("version", usage)

	if buildDate != "" {
		localVersion += ("+" + buildDate)
	}
}

// Do 执行子命令
func Do(output *os.File) error {
	flagset := flag.NewFlagSet("version", flag.ExitOnError)
	check := flagset.Bool("c", false, "是否检测线上的最新版本")
	if err := flagset.Parse(os.Args[1:]); err != nil {
		return err
	}

	if *check {
		return checkRemoteVersion(output)
	}

	_, err := fmt.Fprintf(output, "web:%s build with %s\n", localVersion, runtime.Version())
	return err
}

// 检测框架的最新版本号
//
// 获取线上的标签列表，拿到其中的最大值。
func checkRemoteVersion(output *os.File) error {
	cmd := exec.Command("git", "ls-remote", "--tags")
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
	vers := make([]*version.SemVersion, 0, 10)

	for s.Scan() {
		text := s.Text()
		index := strings.LastIndex(text, "/v")
		if index < 0 {
			continue
		}
		text = text[index+2:]

		ver, err := version.SemVer(text)
		if err != nil {
			return "", err
		}
		vers = append(vers, ver)
	}

	sort.SliceStable(vers, func(i, j int) bool {
		return vers[i].Compare(vers[j]) > 0
	})

	return vers[0].String(), nil
}

func usage(output *os.File) {
	fmt.Fprintln(output, `显示当前程序的版本号

语法：web version [-c]
如果指定了 -c，则会检测是否存在新版本的内容。`)
}
