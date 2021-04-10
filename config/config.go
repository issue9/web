// SPDX-License-Identifier: MIT

// Package config 提供了加载配置项内容的各类方法
package config

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/fs"
	"reflect"

	"gopkg.in/yaml.v2"
)

// Error 表示配置内容字段错误
type Error struct {
	Config, Field, Message string
	Value                  interface{}
}

// UnmarshalFunc 定义了配置项的解码函数原型
//
// config 为配置项的配置内容，可以是一个文件名或是数据库的 DSN 等，
// UnmarshalFunc 负责将其指向的内容解析并映身到 v 对象。
type UnmarshalFunc func(config string, v interface{}) error

// Refresher 带有可刷新功能的配置项管理工具
type Refresher struct {
	config    string
	rValue    reflect.Value
	rType     reflect.Type
	unmarshal UnmarshalFunc
}

// Load 加载配置内容 config 到 v 对象中
//
// config 表示配置项的配置内容，比如文件名、SQL 的 DSN 等；
// v 配置项导出对象的实例指针，最终 config 指向的数据会被解析到 v 中;
// f 从 config 导入到 v 的实现方法；
func Load(config string, v interface{}, f UnmarshalFunc) (*Refresher, error) {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	ref := &Refresher{
		config:    config,
		rValue:    rv,
		rType:     rv.Type(),
		unmarshal: f,
	}
	if err := ref.Refresh(); err != nil {
		return nil, err
	}
	return ref, nil
}

// Refresh 刷新与当前对象关联的配置项对象
func (l *Refresher) Refresh() error {
	v := reflect.New(l.rType)
	if err := l.unmarshal(l.config, v.Interface()); err != nil {
		return err
	}
	l.rValue.Set(v.Elem())

	return nil
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s:%s[%s]", err.Config, err.Field, err.Message)
}

// LoadYAML 加载 YAML 的配置文件并转换成 v 对象的内容
func LoadYAML(f fs.FS) UnmarshalFunc {
	return func(path string, v interface{}) error {
		data, err := fs.ReadFile(f, path)
		if err != nil {
			return err
		}

		return yaml.Unmarshal(data, v)
	}
}

// LoadJSON 加载 JSON 的配置文件并转换成 v 对象的内容
func LoadJSON(f fs.FS) UnmarshalFunc {
	return func(path string, v interface{}) error {
		data, err := fs.ReadFile(f, path)
		if err != nil {
			return err
		}

		return json.Unmarshal(data, v)
	}
}

// LoadXML 加载 XML 的配置文件并转换成 v 对象的内容
func LoadXML(f fs.FS) UnmarshalFunc {
	return func(path string, v interface{}) error {
		data, err := fs.ReadFile(f, path)
		if err != nil {
			return err
		}

		return xml.Unmarshal(data, v)
	}
}
