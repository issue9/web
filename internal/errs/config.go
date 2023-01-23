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
func NewConfigError(field string, msg any) *ConfigError {
	if err, ok := msg.(*ConfigError); ok { // 如果是 ConfigError 类型，那么此操作仅修改此类型的 Field 值
		err.AddFieldParent(field)
		return err
	}

	return &ConfigError{Field: field, Message: msg}
}

// AddFieldParent 为字段名加上一个前缀
//
// 当字段名存在层级关系时，外层在处理错误时，需要为其加上当前层的字段名作为前缀。
func (err *ConfigError) AddFieldParent(prefix string) *ConfigError {
	if prefix == "" {
		return err
	}

	if err.Field == "" {
		err.Field = prefix
		return err
	}

	err.Field = prefix + "." + err.Field
	return err
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
