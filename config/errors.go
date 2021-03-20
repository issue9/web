// SPDX-License-Identifier: MIT

package config

import "fmt"

// Error 表示配置内容字段错误
type Error struct {
	Config, Field, Message string
	Value                  interface{}
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s:%s[%s]", err.Config, err.Field, err.Message)
}
