// SPDX-License-Identifier: MIT

// Package config 提供了从配置文件初始化 Server 的方法
package config

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/fs"

	"gopkg.in/yaml.v2"
)

// Error 表示配置内容字段错误
type Error struct {
	Config, Field, Message string
	Value                  interface{}
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s:%s[%s]", err.Config, err.Field, err.Message)
}

// LoadYAML 加载 YAML 的配置文件并转换成 v 对象的内容
func LoadYAML(f fs.FS, path string, v interface{}) error {
	data, err := fs.ReadFile(f, path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, v)
}

// LoadJSON 加载 JSON 的配置文件并转换成 v 对象的内容
func LoadJSON(f fs.FS, path string, v interface{}) error {
	data, err := fs.ReadFile(f, path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// LoadXML 加载 XML 的配置文件并转换成 v 对象的内容
func LoadXML(f fs.FS, path string, v interface{}) error {
	data, err := fs.ReadFile(f, path)
	if err != nil {
		return err
	}

	return xml.Unmarshal(data, v)
}
