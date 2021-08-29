// SPDX-License-Identifier: MIT

// Package config 提供了从配置文件初始化 server.Options 的方法
package config

import (
	"fmt"

	"golang.org/x/text/message"
)

// Error 表示配置内容字段错误
type Error struct {
	Config  string      // 配置文件的路径
	Field   string      // 字段名
	Message string      // 错误信息
	Value   interface{} // 原始值
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s:%s[%s]", err.Config, err.Field, err.Message)
}

func (err *Error) LocaleString(p *message.Printer) string {
	return p.Sprintf("config error", err.Config, err.Field, err.Message)
}
