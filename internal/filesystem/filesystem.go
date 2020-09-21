// SPDX-License-Identifier: MIT

// Package filesystem 提供与文件系统相关的操作
package filesystem

import "os"

// Exists 文件或是文件夹是否存在
func Exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil || os.IsExist(err)
}
