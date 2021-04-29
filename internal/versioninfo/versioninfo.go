// SPDX-License-Identifier: MIT

// Package versioninfo 提供对版本信息的一些操作
package versioninfo

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/issue9/source"
	"github.com/issue9/version"

	"github.com/issue9/web/internal/filesystem"
)

const (
	//  指定版本化文件的路径
	infoPath    = "internal/version/info.go"
	versionPath = "internal/version/VERSION"
)

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

// DumpInfoFile 输出处理版本信息的 Go 代码
func (dir Dir) DumpInfoFile() error {
	p := filepath.Join(string(dir), infoPath)
	if err := os.MkdirAll(filepath.Dir(p), os.ModePerm); err != nil {
		return err
	}
	return source.DumpGoSource(p, []byte(versionGo))
}

// DumpVersionFile 输出版本文件
//
// wd 表示工作目录，主要用于确定 git 的相关信息。
func (dir Dir) DumpVersionFile(wd string) error {
	cmd := exec.Command("git", "describe", "--abbrev=40", "--tags", "--long")
	cmd.Dir = wd
	buf := new(bytes.Buffer)
	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		return err
	}

	// bs 格式：v0.2.4-0-ge2f5e99a3306bba28e81f507bf66c905825184e5
	// 替换其中的第一个 - 为日期，第二个 -g 为点
	ver := buf.String()
	date := time.Now().Format("+20060102.")
	ver = strings.Replace(ver, "-", date, 1)
	ver = strings.Replace(ver, "-g", ".", 1)
	if ver[0] == 'v' || ver[0] == 'V' {
		ver = ver[1:]
	}

	if !version.SemVerValid(ver) {
		return fmt.Errorf("无法生成正确的版本号：%s", ver)
	}

	return os.WriteFile(filepath.Join(string(dir), versionPath), []byte(ver), fs.ModePerm)
}
