// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package dev Development 环境下的功能实现
package dev

import (
	"path/filepath"
	"strings"
)

func Filename(f string) string {
	ext := filepath.Ext(f)
	base := strings.TrimSuffix(f, ext)

	// 用 _development 而不是 .development，防止在文件没有扩展名的情况下改变了文件的扩展名。
	return base + "_development" + ext
}
