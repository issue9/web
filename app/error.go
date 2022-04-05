// SPDX-License-Identifier: MIT

package app

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

// ConfigSanitizer 对配置文件的数据验证和修正接口
type ConfigSanitizer interface {
	SanitizeConfig() *ConfigError
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
