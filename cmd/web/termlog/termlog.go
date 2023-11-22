// SPDX-License-Identifier: MIT

// Package termlog 统一的终端日志类型
package termlog

import (
	"io"

	"github.com/issue9/localeutil"
	"github.com/issue9/web/logs"
)

// New 声明用于终端输出的日志
//
// 如果初始化时出错，则会直接 Panic。
func New(p *localeutil.Printer, out io.Writer) logs.Logs {
	log, err := logs.New(p, &logs.Options{
		Levels:  logs.AllLevels(),
		Handler: logs.NewTermHandler(out, nil),
		Created: logs.NanoLayout,
	})
	if err != nil {
		panic(err)
	}

	return log
}
