// SPDX-License-Identifier: MIT

// Package versioninfo 提供对版本信息的一些操作
package versioninfo

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/issue9/source"

	_ "embed"

	"github.com/issue9/web/internal/filesystem"
)

//  指定版本化文件的路径
const infoPath = "internal/version/info.go"

//go:embed data/version.go
var versionCode string

// Dir 版本信息的相关数据
type Dir string

// Root 获取 curr 所在项目的根目录
//
// 从 curr 向上级查找，直到找到 go.mod 文件为止，返回该文件所在的目录。
func Root(curr string) (Dir, error) {
	path, err := filepath.Abs(curr)
	if err != nil {
		return "", err
	}

	for {
		if filesystem.Exists(filepath.Join(path, "go.mod")) {
			return Dir(path), nil
		}

		p1 := filepath.Dir(path)
		if path == p1 {
			return "", errors.New("未找到根目录")
		}
		path = p1
	}
}

// DumpFile 输出版本文件
func (dir Dir) DumpFile() error {
	const dateFormat = "20060102"

	cmd := exec.Command("git", "describe", "--abbrev=40", "--tags", "--long")
	cmd.Dir = string(dir)
	buf := new(bytes.Buffer)
	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		return err
	}

	tag, commits, hash := parseDescribe(buf.String())
	now := time.Now().Format(dateFormat)
	code := strings.Replace(versionCode, "VERSION", tag, 1)
	code = strings.Replace(code, "HASH", hash, 1)
	code = strings.Replace(code, "5000", commits, 1)
	code = strings.Replace(code, "DATE", now, 1)
	code = strings.Replace(code, "FORMATE", dateFormat, 1)
	return source.DumpGoSource(filepath.Join(string(dir), infoPath), []byte(code))
}

// s 格式：v0.2.4-0-ge2f5e99a3306bba28e81f507bf66c905825184e5
// 替换其中的第一个 - 为日期，第二个 -g 为点
func parseDescribe(s string) (tag, commits, hash string) {
	s = strings.TrimRightFunc(s, func(r rune) bool { return r == '\n' })

	if index := strings.IndexByte(s, '-'); index > 0 {
		tag = s[:index]
		s = s[index+1:]
	}
	if tag[0] == 'v' || tag[0] == 'V' {
		tag = tag[1:]
	}

	if index := strings.IndexByte(s, '-'); index > 0 {
		commits = s[:index]
		hash = s[index+2:] // 去除 -g
	}

	return
}
