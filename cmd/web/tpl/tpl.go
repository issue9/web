// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

// Package tpl 提供新建项目模板的功能
package tpl

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	xcopy "github.com/otiai10/copy"
	"golang.org/x/mod/module"

	"github.com/issue9/web"
)

const (
	title = web.StringPhrase("create new web project")
	usage = web.StringPhrase(`create new web project

usage:

web new [flags] tpl-path [module-path]

flags:
{{flags}}`)

	licenseUsage = web.StringPhrase("select license")
	authorUsage  = web.StringPhrase("set author in file header")
	extUsage     = web.StringPhrase("select exts")
	yearUsage    = web.StringPhrase("set year in file header")
)

func Init(opt *cmdopt.CmdOpt, p *localeutil.Printer) {
	opt.New("new", title.LocaleString(p), usage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		license := fs.String("l", "MIT", licenseUsage.LocaleString(p))
		author := fs.String("a", "", authorUsage.LocaleString(p))
		extsStr := fs.String("e", "", extUsage.LocaleString(p))
		year := fs.String("y", time.Now().Format("2006"), yearUsage.LocaleString(p))

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
			if err != nil {
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

			exts := strings.Split(*extsStr, ",")
			for index, ext := range exts {
				if ext[0] != '.' {
					exts[index] = "." + ext
				}
			}
			if err := replaceFileHeaders(dest, *author, *license, *year, exts); err != nil {
				return err
			}

			return nil
		}
	})
}

func replaceFileHeaders(dir string, author, license, year string, exts []string) error {
	r := strings.NewReplacer(
		"{{author}}", author,
		"{{license}}", license,
		"{{year}}", year,
	)

	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || slices.Index(exts, filepath.Ext(d.Name())) >= 0 {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		data = []byte(r.Replace(string(data)))
		return os.WriteFile(path, data, os.ModePerm)
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

	return xcopy.Copy(out.Dir, dest)
}
