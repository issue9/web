// SPDX-License-Identifier: MIT

// Package config 提供了对动态加载配置项的支持
package config

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// ErrNotFound 未找到与 ID 相关的项
var ErrNotFound = errors.New("id 值已经存在")

// UnmarshalFunc 定义了配置项的解码函数原型
//
// config 为配置项的配置内容，可以是一个文件名或是数据库的 DSN 等，
// UnmarshalFunc 负责将其指向的内容解析并映身到 v 对象。
type UnmarshalFunc func(config string, v interface{}) error

// Config 管理配置项的相关信息
type Config struct {
	unmarshals []*unmarshal
}

type unmarshal struct {
	id        string
	config    string
	v         interface{}
	unmarshal UnmarshalFunc
	notify    func()
}

// Register 注册配置项
//
// id 表示该项的唯一标记，后期如果需要刷新配置项，会用到此值，在不为空的情况下，要求唯一；
// config 表示配置项的配置内容，比如文件名、SQL 的 DSN 等；
// v 配置项导出对象的实例指针，最终 config 指向的数据会被解析到 v 中;
// f 从 config 导入到 v 的实现方法；
// notify 如果后期需要更新数据，则通过此函数通知用户数据已经被更新；
func (cfg *Config) Register(id, config string, v interface{}, f UnmarshalFunc, notify func()) error {
	if u := cfg.findUnmarshal(id); u != nil {
		return fmt.Errorf("id %s 已经存在在", id)
	}

	cfg.unmarshals = append(cfg.unmarshals, &unmarshal{
		id:        id,
		config:    config,
		v:         v,
		unmarshal: f,
		notify:    notify,
	})
	return nil
}

func (cfg *Config) findUnmarshal(id string) *unmarshal {
	if id != "" {
		for _, u := range cfg.unmarshals {
			if u.id == id {
				return u
			}
		}
	}
	return nil
}

// Refresh 刷新指定 ID 的配置项
func (cfg *Config) Refresh(id string) error {
	if id == "" {
		panic("参数 id 不能为空")
	}

	u := cfg.findUnmarshal(id)
	if u == nil {
		return ErrNotFound
	}

	if err := u.unmarshal(u.config, u.v); err != nil {
		return err
	}

	if u.notify != nil {
		u.notify()
	}

	return nil
}

// LoadYAML 加载 YAML 的配置文件并转换成 v 对象的内容
func LoadYAML(path string, v interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, v)
}

// LoadJSON 加载 JSON 的配置文件并转换成 v 对象的内容
func LoadJSON(path string, v interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// LoadXML 加载 XML 的配置文件并转换成 v 对象的内容
func LoadXML(path string, v interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return xml.Unmarshal(data, v)
}
