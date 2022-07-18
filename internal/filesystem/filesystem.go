// SPDX-License-Identifier: MIT

// Package filesystem 文件系统相关操作
package filesystem

import (
	"errors"
	"io/fs"
	"os"
)

func Exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil || errors.Is(err, fs.ErrExist)
}

func ExistsFS(fsys fs.FS, p string) bool {
	_, err := fs.Stat(fsys, p)
	return err == nil || errors.Is(err, fs.ErrExist)
}
