// SPDX-License-Identifier: MIT

// Package filesystem 提供文件系统的相关操作
package filesystem

import (
	"io/fs"
	"os"
)

func existsFS(fsys fs.FS, p string) bool {
	_, err := fs.Stat(fsys, p)
	return err == nil || os.IsExist(err)
}
