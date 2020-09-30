// SPDX-License-Identifier: MIT

// Package config 提供了加载配置项内容的各类方法
package config

import (
	"errors"
	"reflect"
)

// ErrNotFound 未找到与 ID 相关的项
var ErrNotFound = errors.New("id 值已经存在")

// Config 管理配置项的加载和刷新
type Config struct {
	config    string
	rValue    reflect.Value
	rType     reflect.Type
	unmarshal UnmarshalFunc
	notify    func()
}

// New 注册配置项
//
// config 表示配置项的配置内容，比如文件名、SQL 的 DSN 等；
// v 配置项导出对象的实例指针，最终 config 指向的数据会被解析到 v 中;
// f 从 config 导入到 v 的实现方法；
// notify 如果后期需要更新数据，则通过此函数通知用户数据已经被更新；
func New(config string, v interface{}, f UnmarshalFunc, notify func()) *Config {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	return &Config{
		config:    config,
		rValue:    rv,
		rType:     rv.Type(),
		unmarshal: f,
		notify:    notify,
	}
}

// Refresh 刷新指定 ID 的配置项
func (cfg *Config) Refresh() error {
	v := reflect.New(cfg.rType)
	if err := cfg.unmarshal(cfg.config, v.Interface()); err != nil {
		return err
	}
	cfg.rValue.Set(v.Elem())

	if cfg.notify != nil {
		cfg.notify()
	}

	return nil
}
