// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

// Package tpl 提供新建项目模板的功能
package tpl

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	xcopy "github.com/otiai10/copy"
	"golang.org/x/mod/module"

	"github.com/issue9/web"
)

const modePerm = os.ModePerm

const (
	title = web.StringPhrase("create new web project")
	usage = web.StringPhrase(`create new web project

usage:

web new [flags] tpl-path module-path

flags:
{{flags}}`)

	extUsage        = web.StringPhrase("select exts")
	fileHeaderUsage = web.StringPhrase("set file header")
)

func Init(opt *cmdopt.CmdOpt, p *localeutil.Printer) {
	opt.New("new", title.LocaleString(p), usage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		extsStr := fs.String("e", "", extUsage.LocaleString(p))
		header := fs.String("h", "", fileHeaderUsage.LocaleString(p))

		return func(w io.Writer) error {
			if fs.NArg() != 2 {
				return web.NewLocaleError("tpl-path or module-path not set")
			}

			tplPath := fs.Arg(0) // 模板的获取路径
			dstPath := fs.Arg(1) // 新项目的模块路径
			if err := module.CheckPath(dstPath); err != nil {
				return err
			}
			name := filepath.Base(dstPath) // 保存的目录

			// 检测保存目录的状态
			dest := filepath.Join(".", name)
			de, err := os.ReadDir(dest)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if len(de) > 0 {
				return web.NewLocaleError("dest is not empty")
			}

			// 获取内容
			if err = download(tplPath, dest); err != nil {
				return err
			}

			if err := replaceGo(dest, tplPath, dstPath); err != nil {
				return err
			}

			return insertFileHeaders(dest, *header, *extsStr)
		}
	})
}

// 下载 tplPath 指向的仓库内容并保存在 dest 中
func download(tplPath, dest string) error {
	if !strings.Contains(tplPath, "@") {
		tplPath += "@latest"
	}

	modPath, _, _ := strings.Cut(tplPath, "@")
	if err := module.CheckPath(modPath); err != nil {
		return err
	}

	var stderr, stdout bytes.Buffer
	cmd := exec.Command("go", "mod", "download", "-json", tplPath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	out := &struct{ Dir string }{}
	if err := json.Unmarshal(stdout.Bytes(), out); err != nil {
		return err
	}

	return xcopy.Copy(out.Dir, dest, xcopy.Options{PermissionControl: xcopy.AddPermission(modePerm)})
}
