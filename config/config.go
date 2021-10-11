// SPDX-License-Identifier: MIT

// Package config 提供了从配置文件初始化 server.Options 的方法
package config

import (
	"github.com/issue9/localeutil"
	"golang.org/x/text/message"
)

// Error 表示配置内容字段错误
type Error struct {
	Config  string      // 配置文件的路径
	Field   string      // 字段名
	Message interface{} // 错误信息
	Value   interface{} // 原始值
}

func (err *Error) Error() string {
	return err.LocaleString(localeutil.EmptyPrinter())
}

func (err *Error) LocaleString(p *message.Printer) string {
	msg := err.Message
	if ls, ok := err.Message.(localeutil.LocaleStringer); ok {
		msg = ls.LocaleString(p)
	}

	return p.Sprintf("config error", err.Config, err.Field, msg)
}
