// SPDX-License-Identifier: MIT

package errs

import (
	"fmt"

	"github.com/issue9/localeutil"
	"golang.org/x/text/message"
)

// ConfigError 表示配置内容字段错误
type ConfigError struct {
	Path    string // 配置文件的路径
	Field   string // 字段名
	Message any    // 错误信息
	Value   any    // 字段的原始值
}

// NewConfigError 返回表示配置文件错误的对象
//
// field 表示错误的字段名；
// msg 表示错误信息，可以是任意类型；
// path 表示配置文件的路径；
// val 表示错误字段的原始值；
func NewConfigError(field string, msg any, path string, val any) *ConfigError {
	return &ConfigError{
		Field:   field,
		Message: msg,
		Path:    path,
		Value:   val,
	}
}

func (err *ConfigError) Error() string {
	var msg string
	switch inst := err.Message.(type) {
	case fmt.Stringer:
		msg = inst.String()
	case error:
		msg = inst.Error()
	default:
		msg = fmt.Sprint(err.Message)
	}

	return fmt.Sprintf("%s at %s:%s", msg, err.Path, err.Field)
}

func (err *ConfigError) LocaleString(p *message.Printer) string {
	msg := err.Message
	if ls, ok := err.Message.(localeutil.LocaleStringer); ok {
		msg = ls.LocaleString(p)
	}

	return p.Sprintf("%s at %s:%s", msg, err.Path, err.Field)
}
