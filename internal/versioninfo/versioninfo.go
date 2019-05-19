// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package versioninfo 提供对版本信息的一些操作
package versioninfo

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/issue9/utils"
)

// Path 指定版本化文件的路径
const Path = "internal/version/version.go"

// FindRoot 查找项目的根目录
//
// 从当前目录开始向上查找，以找到 go.mod 为准，如果没有 go.mod 则会失败。
func FindRoot(curr string) (string, error) {
	path, err := filepath.Abs(curr)
	if err != nil {
		return "", err
	}

	for {
		if utils.FileExists(filepath.Join(path, "go.mod")) {
			return path, nil
		}

		p1 := filepath.Dir(path)
		if path == p1 {
			return "", errors.New("未找到根目录")
		}
		path = p1
	}
}

// DumpFile 输出版本信息文件
//
// ver 为需要指定的版本号
func DumpFile(ver string) error {
	root, err := FindRoot("./")
	if err != nil {
		return err
	}

	p := filepath.Join(root, Path)
	if err := os.MkdirAll(filepath.Dir(p), os.ModePerm); err != nil {
		return err
	}

	return utils.DumpGoFile(p, fmt.Sprintf(versiongo, ver))
}

// VarPath 获取 buildDate 的路径
func VarPath() (string, error) {
	p, err := FindRoot("./")
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadFile(filepath.Join(p, "/go.mod"))
	if err != nil {
		return "", err
	}

	s := bufio.NewScanner(bytes.NewBuffer(data))
	s.Split(bufio.ScanLines)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "module ") {
			return path.Join(line[len("module "):], Path), nil
		}
	}

	return "", errors.New("未找到模块名称")
}
