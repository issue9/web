// SPDX-License-Identifier: MIT

// Package config 提供了从配置文件初始化 server.Options 的方法
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
