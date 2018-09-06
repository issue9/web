// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package config 提供了对多种格式配置文件的支持
package config

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

var errExtNotAllowEmpty = errors.New("扩展名不能为空")

// UnmarshalFunc 定义了将文本内容解析到对象的函数原型。
type UnmarshalFunc func([]byte, interface{}) error

// Sanitizer 如果对象实现了该方法，那么在解析完之后，
// 会调用该接口的函数对数据进行修正和检测。
type Sanitizer interface {
	Sanitize() error
}

var unmarshals = map[string]UnmarshalFunc{}

func init() {
	if err := AddUnmarshal(json.Unmarshal, "json"); err != nil {
		panic(err)
	}

	if err := AddUnmarshal(yaml.Unmarshal, "yaml", ".yml"); err != nil {
		panic(err)
	}

	if err := AddUnmarshal(xml.Unmarshal, "xml"); err != nil {
		panic(err)
	}
}

// AddUnmarshal 注册解析函数
func AddUnmarshal(m UnmarshalFunc, ext ...string) error {
	for _, e := range ext {
		if e == "" || e == "." {
			return errExtNotAllowEmpty
		}

		if e[0] != '.' {
			e = "." + e
		}

		e = strings.ToLower(e)
		if _, found := unmarshals[e]; found {
			return fmt.Errorf("已经存在该扩展名 %s 对应的解析函数", ext)
		}
		unmarshals[e] = m
	}

	return nil
}

// SetUnmarshal 修改指定扩展名关联的解析函数，不存在则添加。
func SetUnmarshal(m UnmarshalFunc, ext ...string) error {
	for _, e := range ext {
		if e == "" || e == "." {
			return errExtNotAllowEmpty
		}

		if e[0] != '.' {
			e = "." + e
		}

		unmarshals[strings.ToLower(e)] = m
	}

	return nil
}

// LoadFile 加载指定的配置文件内容到 v 中
func LoadFile(path string, v interface{}) error {
	r, err := os.Open(path)
	if err != nil {
		return err
	}
	return Load(r, filepath.Ext(path), v)
}

// Load 加载指定的配置文件内容到 v 中
func Load(r io.Reader, typ string, v interface{}) error {
	typ = strings.ToLower(typ)
	unmarshal, found := unmarshals[typ]
	if !found {
		return fmt.Errorf("无效的配置文件类型：%s", typ)
	}

	data, err := ioutil.ReadAll(r)
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
