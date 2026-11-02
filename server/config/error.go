// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"

	"github.com/issue9/localeutil"

	"github.com/issue9/web"
	"github.com/issue9/web/filter"
)

type ConfigFilter interface {
	FilterConfig(*filter.Filter)
}

// 表示配置内容字段错误
type fieldError struct {
	Path    string // 配置文件的路径
	Name    string
	Message web.LocaleStringer
}

func (f *fieldError) collect(name string, reason web.LocaleStringer) {
	f.Name = name
	f.Message = reason
}

func (err *fieldError) Error() string {
	var msg string
	switch v := err.Message.(type) {
	case fmt.Stringer:
		msg = v.String()
	case error:
		msg = v.Error()
	default:
		msg = fmt.Sprint(err.Message)
	}

	return fmt.Sprintf("%s at %s:%s", msg, err.Path, err.Name)
}

func (err *fieldError) LocaleString(p *localeutil.Printer) string {
	return localeutil.Phrase("%s at %s:%s", err.Message, err.Path, err.Name).LocaleString(p)
}

func newFieldError(field string, msg any) *fieldError {
	if err, ok := msg.(*fieldError); ok {
		err.Name = field + "." + err.Name
		return err
	}
	if m, ok := msg.(localeutil.Stringer); ok {
		return &fieldError{Name: field, Message: m}
	}
	return &fieldError{Name: field, Message: localeutil.Phrase(fmt.Sprint(msg))}
}
