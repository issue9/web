// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package build 编译程，直接引用 go build
package build

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/issue9/cmdopt"

	"github.com/issue9/web/internal/cmd/release"
)

var flagset *flag.FlagSet

// Init 初始化函数
func Init(opt *cmdopt.CmdOpt) {
	flagset = opt.New("build", do, usage)
}

func do(output io.Writer) error {
	args := make([]string, 0, len(flagset.Args())+1)
	args = append(args, "build")
	args = append(args, flagset.Args()...)

	mp, err := getModuleName()
	if err != nil {
		return err
	}
	mp = path.Join(mp, release.Path)
	arg := `"-X ` + mp + ".buildDate=" + time.Now().Format("20060102") + `"`
	args = append(args, "ldflags", arg)

	cmd := exec.Command("go", args...)
	cmd.Stderr = output
	cmd.Stdout = output

	return cmd.Run()
}

func usage(output io.Writer) error {
	_, err := fmt.Fprintln(output, `编译当前程序，功能与 go build 完全相同！

如果你使用 web release 发布了版本号，则当前操作还会在每一次编译时指定个编译日期，
固定格式为 YYYYMMDD。`)
	return err
}

func getModuleName() (string, error) {
	path, err := release.FindRoot("./")
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadFile(filepath.Join(path, "/go.mod"))
	if err != nil {
		return "", err
	}

	s := bufio.NewScanner(bytes.NewBuffer(data))
	s.Split(bufio.ScanLines)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "module ") {
			return line[len("module "):], nil
		}
	}

	return "", errors.New("未找到模块名称")
}
