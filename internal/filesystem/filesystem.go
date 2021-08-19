// SPDX-License-Identifier: MIT

// Package filesystem 提供文件系统的相关操作
package filesystem

import (
	"io/fs"
	"os"
)

func Exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil || os.IsExist(err)
}

func ExistsFS(fsys fs.FS, p string) bool {
	_, err := fs.Stat(fsys, p)
	return err == nil || os.IsExist(err)
}
