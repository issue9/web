// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package config 配置文件的处理
package config

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// UnmarshalFunc 定义从文本内容加载到对象的函数。
type UnmarshalFunc func([]byte, interface{}) error

// Sanitizer 检测配置文件内容
type Sanitizer interface {
	Sanitize() error
}

var unmarshals = map[string]UnmarshalFunc{}

func init() {
	if err := Register(json.Unmarshal, "json"); err != nil {
		panic(err)
	}

	if err := Register(yaml.Unmarshal, "yaml", ".yml"); err != nil {
		panic(err)
	}

	if err := Register(xml.Unmarshal, "xml"); err != nil {
		panic(err)
	}
}

// Register 注册解析函数
func Register(m UnmarshalFunc, ext ...string) error {
	for _, e := range ext {
		if e == "" {
			return errors.New("扩展名不能为空")
		}

		if e[0] != '.' {
			e = "." + e
		}

		if _, found := unmarshals[e]; found {
			return fmt.Errorf("已经存在该扩展名 %s 对应的解析函数", ext)
		}
		unmarshals[e] = m
	}

	return nil
}

// Load 加载指定的配置文件内容到 v 中
func Load(path string, v interface{}) error {
	ext := strings.ToLower(filepath.Ext(path))
	unmarshal, found := unmarshals[ext]
	if !found {
		return fmt.Errorf("无效的配置文件类型：%s", ext)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err = unmarshal(data, v); err != nil {
		return err
	}

	if s, ok := v.(Sanitizer); ok {
		return s.Sanitize()
	}
	return nil
}
