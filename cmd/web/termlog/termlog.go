// SPDX-License-Identifier: MIT

// Package termlog 统一的终端日志类型
package termlog

import (
	"io"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v7"
)

// New 声明用于终端输出的日志
//
// 如果初始化时出错，则会直接 Panic。
func New(p *localeutil.Printer, out io.Writer) *logs.Logs {
	return logs.New(
		logs.NewTermHandler(out, nil),
		logs.WithLevels(logs.AllLevels()...),
		logs.WithCreated(logs.NanoLayout),
	)
}
