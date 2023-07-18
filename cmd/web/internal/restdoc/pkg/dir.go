// SPDX-License-Identifier: MIT

package pkg

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/issue9/sliceutil"
	"golang.org/x/mod/modfile"
)

// 获取 root 以及其子目录列表
func getDirs(root string, recursive bool) ([]string, error) {
	if !recursive {
		return []string{root}, nil
	}

	dirs := make([]string, 0, 100)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			dirs = append(dirs, p)
		}
		return err
	})

	if err != nil {
		return nil, err
	}
	return dirs, nil
}

func getModPath(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	pkgNames := make([]string, 0, 10)
LOOP:
	for {
		p := filepath.Join(abs, "go.mod")
		stat, err := os.Stat(p)
		switch {
		case err == nil:
			if stat.IsDir() { // 名为 go.mod 的目录
				pkgNames = append(pkgNames, stat.Name())
				abs = filepath.Dir(abs)
				continue LOOP
			}

			data, err := os.ReadFile(p)
			if err != nil {
				return "", err
			}
			mod, err := modfile.Parse(p, data, nil)
			if err != nil {
				return "", err
			}

			pkgNames = append(pkgNames, mod.Module.Mod.Path)
			sliceutil.Reverse(pkgNames)
			return path.Join(pkgNames...), nil
		case errors.Is(err, os.ErrNotExist):
			// 这两行不能用 filepath.Split 代替，split 会为 abs1 留下最后的分隔符，
			// 导致下一次的 filepath.Split 返回空的 file 值。
			base := filepath.Base(abs)
			abs1 := filepath.Dir(abs)

			if abs1 == abs { // 到根目录了
				return "", os.ErrNotExist
			}

			abs = abs1
			pkgNames = append(pkgNames, base)
			continue LOOP
		default: // 文件存在，但是出错。
			return "", err
		}
	}
}
