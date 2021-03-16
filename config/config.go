// SPDX-License-Identifier: MIT

// Package config 提供了加载配置项内容的各类方法
package config

import "reflect"

// Loader 配置管理接口
type Loader interface {
	Load(config string, v interface{}) (Refresher, error)
}

// SelectorFunc 根据参数确定使用哪个解码函数
type SelectorFunc func(string) UnmarshalFunc

// UnmarshalFunc 定义了配置项的解码函数原型
//
// config 为配置项的配置内容，可以是一个文件名或是数据库的 DSN 等，
// UnmarshalFunc 负责将其指向的内容解析并映身到 v 对象。
type UnmarshalFunc func(config string, v interface{}) error

// Refresher 带有刷新功能的配置项加载接口
type Refresher interface {
	Refresh() error
}

type refresher struct {
	config    string
	rValue    reflect.Value
	rType     reflect.Type
	unmarshal UnmarshalFunc
}

// Load 加载配置内容到 v 对象中
//
// config 表示配置项的配置内容，比如文件名、SQL 的 DSN 等；
// v 配置项导出对象的实例指针，最终 config 指向的数据会被解析到 v 中;
// f 从 config 导入到 v 的实现方法；
func Load(config string, v interface{}, f UnmarshalFunc) Refresher {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	return &refresher{
		config:    config,
		rValue:    rv,
		rType:     rv.Type(),
		unmarshal: f,
	}
}

// Refresh 刷新指定配置项
func (l *refresher) Refresh() error {
	v := reflect.New(l.rType)
	if err := l.unmarshal(l.config, v.Interface()); err != nil {
		return err
	}
	l.rValue.Set(v.Elem())

	return nil
}
