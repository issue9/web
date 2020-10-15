// SPDX-License-Identifier: MIT

// Package version 显示版本号信息
package version

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/issue9/cmdopt"

	"github.com/issue9/web/internal/version"
)

const usage = `显示当前程序的版本号

语法：web version [options]
`

// 用于获取版本信息的 git 仓库地址
//
// 在测试环境会被修改为当前目录
var repoURL = "https://github.com/issue9/web"

var (
	check   bool
	list    bool
	flagset *flag.FlagSet
)

// Init 初始化函数
func Init(opt *cmdopt.CmdOpt) {
	flagset = opt.New("version", usage, do)
	flagset.BoolVar(&check, "c", false, "是否检测线上的最新版本")
	flagset.BoolVar(&list, "l", false, "显示线上的所有版本号")
}

func do(output io.Writer) error {
	switch {
	case check:
		return checkRemoteVersion(output)
	case list:
		return checkRemoteVersion(output)
	default:
		return printLocalVersion(output)
	}
}

// 检测框架的最新版本号
//
// 获取线上的标签列表，拿到其中的最大值。
func checkRemoteVersion(output io.Writer) error {
	tags, err := getRemoteTags()
	if err != nil {
		return err
	}

	if err = printLocalVersion(output); err != nil {
		return err
	}

	_, err = fmt.Fprintf(output, "latest:%s\n", tags[0])
	return err
}

func printLocalVersion(output io.Writer) error {
	_, err := fmt.Fprintf(output, "web: %s\ngo: %s\n", version.FullVersion(), strings.TrimPrefix(runtime.Version(), "go"))
	return err
}

func getRemoteVersions(output io.Writer) error {
	tags, err := getRemoteTags()
	if err != nil {
		return err
	}

	for _, tag := range tags {
		if _, err = fmt.Fprintln(output, tag); err != nil {
			return err
		}
	}

	return nil
}

// 获取线上的标签列表。
func getRemoteTags() ([]string, error) {
	// 返回以下格式内容
	// a07f91201239035ebf85a6423016a6b736b0d037	refs/tags/v0.16.2
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
		index := strings.LastIndex(text, "/v")
		if index < 0 {
			continue
		}
		tags = append(tags, text[index+2:])
	}

	sort.Sort(sort.Reverse(sort.StringSlice(tags)))

	return tags, nil
}
