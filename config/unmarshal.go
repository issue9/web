// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// UnmarshalFunc 定义了配置项的解码函数原型
//
// config 为配置项的配置内容，可以是一个文件名或是数据库的 DSN 等，
// UnmarshalFunc 负责将其指向的内容解析并映身到 v 对象。
type UnmarshalFunc func(config string, v interface{}) error

var fileUnmarshals = map[string]UnmarshalFunc{
	".yaml": LoadYAML,
	".yml":  LoadYAML,
	".xml":  LoadXML,
	".json": LoadJSON,
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

// LoadFile 加载配置文件内容至 v
//
// 支持 yaml、xml 和 json 格式内容，以后缀名作为判断依据。
func LoadFile(path string, v interface{}) error {
	ext := strings.ToLower(filepath.Ext(path))
	if f, found := fileUnmarshals[ext]; found {
		return f(path, v)
	}

	return errors.New("无法处理的文档类型：" + ext)
}
