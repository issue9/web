// SPDX-License-Identifier: MIT

// 简单的辅助功能命令行工具
package main

import "github.com/issue9/web/internal/cmd"

func main() {
	if err := cmd.Exec(); err != nil {
		panic(err)
	}
}
