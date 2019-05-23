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
	"time"

	"github.com/issue9/utils"
)

// Path 指定版本化文件的路径
const Path = "internal/version/info.go"

const (
	buildDateName   = "buildDate"
	buildDateLayout = "20060102"
)

// VersionInfo 版本信息的相关数据
type VersionInfo struct {
	root string
}

// New 声明 VersionInfo 变量
func New(curr string) (*VersionInfo, error) {
	root, err := findRoot(curr)
	if err != nil {
		return nil, err
	}

	return &VersionInfo{
		root: root,
	}, nil
}

// findRoot 查找项目的根目录
//
// 从当前目录开始向上查找，以找到 go.mod 为准，如果没有 go.mod 则会失败。
func findRoot(curr string) (string, error) {
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

// Path 返回基于当前项目根目录的文件地址
func (v *VersionInfo) Path(p string) string {
	return filepath.Join(v.root, p)
}

// DumpFile 输出版本信息文件
//
// ver 为需要指定的版本号
func (v *VersionInfo) DumpFile(ver string) error {
	p := v.Path(Path)
	if err := os.MkdirAll(filepath.Dir(p), os.ModePerm); err != nil {
		return err
	}

	return utils.DumpGoFile(p, fmt.Sprintf(versiongo, ver, buildDateName))
}

// LDFlags 获取 ldflags 的参数
//
// p 为开始查找的目录；
// 返回格式为：
//  -X xx.buildDate=20060102
func (v *VersionInfo) LDFlags() (string, error) {
	data, err := ioutil.ReadFile(v.Path("go.mod"))
	if err != nil {
		return "", err
	}

	s := bufio.NewScanner(bytes.NewBuffer(data))
	s.Split(bufio.ScanLines)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "module ") {
			p := path.Join(line[len("module "):], Path)
			p = path.Dir(p)

			date := time.Now().Format(buildDateLayout)
			return fmt.Sprintf(`"-X %s.%s=%s"`, p, buildDateName, date), nil
		}
	}

	return "", errors.New("go.mod 中未找到 module 语句")
}
