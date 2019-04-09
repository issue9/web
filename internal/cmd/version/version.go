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
	"io"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/issue9/web"
)

// 用于获取版本信息的 git 仓库地址
//
// 在测试环境会被修改为当前目录
var repoURL = "https://github.com/issue9/web"

var (
	localVersion = web.Version
	buildDate    string
)

var (
	check   bool
	list    bool
	flagset *flag.FlagSet
)

func init() {
	if buildDate != "" {
		localVersion += ("+" + buildDate)
	}

	flagset = flag.NewFlagSet("version", flag.ExitOnError)
	flagset.BoolVar(&check, "c", false, "是否检测线上的最新版本")
	flagset.BoolVar(&list, "l", false, "显示所有版本号")
}

// Do 执行子命令
func Do(output io.Writer) error {
	if err := flagset.Parse(os.Args[2:]); err != nil {
		return err
	}

	if check {
		return checkRemoteVersion(output)
	}

	if list {
		return getRemoteVersions(output)
	}

	_, err := fmt.Fprintf(output, "web:%s build with %s\n", localVersion, runtime.Version())
	return err
}

// 检测框架的最新版本号
//
// 获取线上的标签列表，拿到其中的最大值。
func checkRemoteVersion(output io.Writer) error {
	tags, err := getRemoteTags()
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(output, "local:%s build with %s\n", localVersion, runtime.Version())
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(output, "latest:%s\n", tags[0][1:])
	return err
}

func getRemoteVersions(output io.Writer) error {
	tags, err := getRemoteTags()
	if err != nil {
		return err
	}

	for _, tag := range tags {
		if _, err = fmt.Fprintf(output, tag); err != nil {
			return err
		}
	}

	return nil
}

// 获取线上的标签列表。
func getRemoteTags() ([]string, error) {
	cmd := exec.Command("git", "ls-remote", "--tags", repoURL)
	buf := new(bytes.Buffer)
	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	tags := make([]string, 0, 100)
	s := bufio.NewScanner(buf)

	for s.Scan() {
		text := s.Text()
		index := strings.LastIndex(text, "/")
		if index < 0 {
			continue
		}
		tags = append(tags, text[index+1:])
	}

	sort.Sort(sort.Reverse(sort.StringSlice(tags)))

	return tags, nil
}

// Usage 当前子命令的用法
func Usage(output io.Writer) {
	fmt.Fprintln(output, `显示当前程序的版本号

语法：web version [options]
options`)
	flagset.SetOutput(output)
	flagset.PrintDefaults()
}
