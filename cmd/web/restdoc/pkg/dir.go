// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package pkg

import (
	"io/fs"
	"os"
	"path/filepath"
)

// 获取 root 以及其子目录列表
func getDirs(root string, recursive bool) ([]string, error) {
	if !recursive {
		return []string{root}, nil
	}

	dirs := make([]string, 0, 100)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			// 忽略目录中的子模块
			if stat, err := os.Stat(filepath.Join(p, "go.mod")); err == nil && !stat.IsDir() && p != root {
				return fs.SkipDir
			}
			dirs = append(dirs, p)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return dirs, nil
}
