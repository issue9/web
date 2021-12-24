// SPDX-License-Identifier: MIT

package app

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

// Sanitizer 自定义配置文件格式需要实现的接口
type Sanitizer interface {
	// Sanitize 对整个配置对象内容的检测
	Sanitize() *Error
}

type EmptyData struct{}

func (err *Error) Error() string {
	return err.LocaleString(localeutil.EmptyPrinter())
}

func (err *Error) LocaleString(p *message.Printer) string {
	msg := err.Message
	if ls, ok := err.Message.(localeutil.LocaleStringer); ok {
		msg = ls.LocaleString(p)
	}

	return p.Sprintf("%s at %s:%s", msg, err.Config, err.Field)
}

func (d *EmptyData) Sanitize() *Error { return nil }
