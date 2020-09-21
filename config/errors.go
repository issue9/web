// SPDX-License-Identifier: MIT

package config

// FieldError 表示配置内容字段错误
type FieldError struct {
	Field, Message string
}

func (err *FieldError) Error() string {
	return err.Message
}
